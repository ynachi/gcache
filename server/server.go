package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/ynachi/gcache/command"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
	"github.com/ynachi/gcache/gerror"
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
	cache    *db.Cache
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
func NewServer(ip string, port int, logLevel string, maxItems int64, evictionPolicyName string) (*Server, error) {
	connString := fmt.Sprintf("%s:%d", ip, port)
	listener, err := net.Listen("tcp", connString)
	if err != nil {
		return nil, err
	}

	cache, err := db.NewCache(maxItems, evictionPolicyName)
	if err != nil {
		return nil, err
	}

	server := &Server{
		address:  listener.Addr().String(),
		listener: listener,
		cache:    cache,
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

// Start starts the server. It listens to new connections and processes them.
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
			s.logger.Debug("gracefully shutdown server")
			if err := s.listener.Close(); err != nil {
				s.logger.Error("error closing listener", "error", err)
			}
			return
		case conn, ok := <-newConns:
			// NoK means newConns channel is closed.
			// So drop because we would not be able to process connections, anyway
			if !ok {
				s.logger.Debug("connections channel was closed")
				return
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
			newConns <- MakeConnection(c, s.cache)
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
			cmd, err := conn.GetCommand()
			if err == nil {
				// process unknown command
				if _, ok := cmd.(*command.Unknown); ok {
					s.SendError(gerror.ErrInvalidCmdName.Error(), conn.writer)
					continue
				}

				// Apply command
				s.logger.Debug("command received", "client_ip", conn.clientIP, "cmd", cmd.Name())
				cmd.Apply(conn.storage, conn.writer)
			}

			// Exit on IOF. Log network unavailability ones to the client. Send the rest to the client.
			if s.handleConnectionError(conn, err) {
				return
			}
		}
	}
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

// handleConnectionError handles connection errors and tells if the caller should return.
// In all other cases, the caller should continue the execution as those are temporary.
func (s *Server) handleConnectionError(conn Connection, err error) (shouldExit bool) {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		s.logger.Debug("client initiated shutdown", "client_ip", conn.clientIP)
		return true
	}

	// Also check for other network errors
	nErr, ok := err.(net.Error)
	if ok && nErr.Timeout() {
		s.logger.Error("network timeout", "client_ip", conn.clientIP, "err", err)
		return false
	}
	if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "read: connection reset by peer" {
		s.logger.Error("Connection reset by peer.", "client_ip", conn.clientIP, "err", err)
		return false
	}

	s.logger.Error("error while handling command", "client_ip", conn.clientIP, "err", err)
	s.SendError(err.Error(), conn.writer)
	return false
}
