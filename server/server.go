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

const (
	LevelDebug = "DEBUG"
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
			// Get command first
			cmd, err := GetCommand(conn.reader)
			if err != nil {
				// EOF means the client is done, so exit.
				if errors.Is(err, io.EOF) {
					s.logger.Debug("client initiated shutdown", "client_ip", conn.clientIP)
					return
				}
				s.logger.Error("error while handling command", "client_ip", conn.clientIP, "err", err)
				s.SendError(err.Error(), conn.writer)
				continue
			}

			// Apply command
			s.logger.Debug("command received", "client_ip", conn.clientIP, "cmd", cmd.String())
			cmd.Apply(nil, conn.writer)
		}
	}
}

// GetCommand handles a command received by the server over an established connection.
func GetCommand(r *bufio.Reader) (command.Command, error) {
	cmdFrameArray, err := GetFrameArray(r)
	if err != nil {
		return nil, err
	}
	return parseCommandFromFrame(cmdFrameArray)
}

// parseCommandFromFrame extracts a command from a frame array.
func parseCommandFromFrame(f *frame.Array) (command.Command, error) {
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

var ErrNotAGcacheCommand = errors.New("command should be an Array frame")

func GetFrameArray(r *bufio.Reader) (*frame.Array, error) {
	cmdFrame, err := frame.Decode(r)
	if err != nil {
		return nil, err
	}
	arrayFrame, ok := cmdFrame.(*frame.Array)
	if !ok {
		return nil, ErrNotAGcacheCommand
	}
	return arrayFrame, nil
}

// SendError responds to a client with an error.
// The error message should be compatible with RESP Error type (i.e., Simple String).
func (s *Server) SendError(msg string, w *bufio.Writer) {
	defer func(conn *bufio.Writer) {
		if err := conn.Flush(); err != nil {
			s.logger.Error("failed to flush buffer to writer", "error", err)
		}
	}(w)
	errFrame, err := frame.NewError(msg)
	if err != nil {
		s.logger.Error("error creating error frame", "error", err)
	}
	_, err = errFrame.WriteTo(w)
	if err != nil {
		s.logger.Error("error writing error frame to network", "error", err)
	}
}
