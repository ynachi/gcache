package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/ynachi/gcache/command"
	"github.com/ynachi/gcache/frame"
	"io"
	"log/slog"
	"net"
	"os"
	"time"
)

type Server struct {
	address  string
	listener net.Listener
	logger   *slog.Logger
}

// Connection is a helper struct which help propagates embedded the treader and writer of a connection while
// allowing top propagates information about this connection.
type Connection struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	clientIP string
}

// MakeConnection creates a connection from a net.Conn object.
func MakeConnection(c net.Conn) Connection {
	return Connection{
		reader:   bufio.NewReader(c),
		writer:   bufio.NewWriter(c),
		clientIP: c.RemoteAddr().String(),
		conn:     c,
	}
}

func (c Connection) Close() error {
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return c.conn.Close()
}

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

// getLogLevel associates a log level to a string.
func getLogLevel(level string) slog.Level {
	switch level {
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelDebug:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

// NewServer creates a new Server with the provided IP address and port.
// It starts listening for incoming connections on the specified address and port.
// Returns a pointer to the created Server or an error if the listener fails to start.
func NewServer(ip string, port int, logLevel string) (*Server, error) {
	connString := fmt.Sprintf("%s:%d", ip, port)
	listener, err := net.Listen("tcp", connString)
	if err != nil {
		return nil, err
	}
	server := &Server{
		address:  listener.Addr().String(),
		listener: listener,
	}
	server.setLogger(logLevel)
	return server, nil
}

// setLogger configures a logger for the server.
func (s *Server) setLogger(level string) {
	opts := slog.HandlerOptions{Level: getLogLevel(level)}
	handler := slog.NewJSONHandler(os.Stdout, &opts)
	s.logger = slog.New(handler)
}

func (s *Server) Address() string {
	return s.address
}

// Start starts the server. It initiates connections handling and command processing.
func (s *Server) Start(ctx context.Context) {
	defer func() {
		if err := s.listener.Close(); err != nil {
			s.logger.Error("error closing listener", "error", err)
		}
	}()

	newConns := make(chan Connection)
	go s.listen(ctx, newConns)

	for {
		select {
		case <-ctx.Done():
			// Close all connections
			s.logger.Info("gracefully shutdown server")
			_ = s.listener.Close()
			return
		case conn, ok := <-newConns:
			// NoK means newConns channel is closed.
			// So drop because we would not be able to process connections, anyway
			if !ok {
				s.logger.Debug("connections channel was closed")
			}
			go s.handleConnection(ctx, conn)
		}
	}
}

// listen waits for new connections for the lifetime of the server.
func (s *Server) listen(ctx context.Context, newConns chan<- Connection) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c, err := s.listener.Accept()
			if err != nil {
				s.logger.Error("error accepting connection", "error", err)
				// TODO: implement exponential backoff later
				time.Sleep(5 * time.Second)
				continue
			}
			newConns <- MakeConnection(c)
		}
	}
}

// attemptCloseConnection tries to close a connection and log an error if it cannot.
func (s *Server) attemptCloseConnection(conn Connection) {
	if err := conn.Close(); err != nil {
		s.logger.Error("error closing connection", "error", err)
	}
}

// handleConnection is the starting point of each connection established with the server.
// It reads command from the connection, apply them and send the response back to the client.
func (s *Server) handleConnection(ctx context.Context, conn Connection) {
	defer s.attemptCloseConnection(conn)
	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("initiating graceful termination", "client_ip", conn.clientIP)
			return
		default:
			// Commands are sent through Array type frame.
			// Read and process them.
			s.handleCommand(conn)
		}
	}
}

// handleCommand handles a command received by the server over an established connection.
func (s *Server) handleCommand(conn Connection) {
	var ErrNotAGcacheCommand = errors.New("command should be an Array frame")
	cmdFrame, err := frame.Decode(conn.reader)
	if err != nil {
		s.handleError(conn, err, err.Error(), "")
		return
	}
	arrayFrame, ok := cmdFrame.(*frame.Array)
	if !ok {
		s.handleError(conn, ErrNotAGcacheCommand, ErrNotAGcacheCommand.Error(), "")
		return
	}
	s.logger.Debug("command received", "client_ip", conn.clientIP, "cmd", arrayFrame.String())
	cmd, err := s.getCommand(arrayFrame)
	if err != nil {
		s.handleError(conn, err, err.Error(), arrayFrame.String())
		return
	}
	cmd.Apply(nil, conn.writer)
}

// getCommand extracts a command from a frame array.
func (s *Server) getCommand(f *frame.Array) (command.Command, error) {
	cmdName, err := command.GetCmdName(f)
	if err != nil {
		return nil, err
	}
	cmd := command.NewCommand(cmdName)
	err = cmd.FromFrame(f)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// handleError is a helper process errors and send feedbacks to the client if needed.
// message is used to send a meaningful error message to the client instead of the original error message.
// EOF means the client is done. We don't need to send feedback in such cases.
func (s *Server) handleError(conn Connection, err error, message string, command string) {
	if errors.Is(err, io.EOF) {
		s.logger.Debug("client initiated shutdown", "client_ip", conn.clientIP)
	} else {
		s.logger.Error(message, "client_ip", conn.clientIP, "cmd", command)
		s.SendError(message, conn.writer)
	}
}

// SendError responds to a client with an error.
// The error message should be compatible to RESP Error type (i.e. Simple String).
func (s *Server) SendError(msg string, conn *bufio.Writer) {
	defer func(conn *bufio.Writer) {
		if err := conn.Flush(); err != nil {
			s.logger.Error("failed to flush buffer to writer", "error", err)
		}
	}(conn)
	errFrame, err := frame.NewError(msg)
	if err != nil {
		s.logger.Error("error creating error frame", "error", err)
	}
	_, err = errFrame.WriteTo(conn)
	if err != nil {
		s.logger.Error("error writing error frame to network", "error", err)
	}
}
