package server

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ynachi/gcache/frame"
	"io"
	"log/slog"
	"net"
	"os"
	"time"
)

type Server struct {
	ip       string
	port     int
	listener net.Listener
	logger   *slog.Logger
}

type LogLevel int

// getLogLevel associates a log level to a string.
func getLogLevel(level string) slog.Level {
	switch level {
	case "WARN", "Warn", "warn", "Warning", "WARNING":
		return slog.LevelWarn
	case "ERROR", "Error":
		return slog.LevelError
	case "DEBUG", "Debug":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

// New creates a new Server with the provided IP address and port.
// It starts listening for incoming connections on the specified address and port.
// Returns a pointer to the created Server or an error if the listener fails to start.
func New(ip string, port int, logLevel string) (*Server, error) {
	connString := fmt.Sprintf("%s:%d", ip, port)
	listener, err := net.Listen("tcp", connString)
	if err != nil {
		return nil, err
	}
	return &Server{
		ip:       ip,
		port:     port,
		listener: listener,
		logger:   newLogger(logLevel),
	}, nil
}

// newLogger returns a new logger with default level.
// The level is typically set via environment variable.
func newLogger(level string) *slog.Logger {
	opts := slog.HandlerOptions{Level: getLogLevel(level)}
	handler := slog.NewJSONHandler(os.Stdout, &opts)
	return slog.New(handler)
}

func (s *Server) Start(ctx context.Context) {
	if s == nil {
		_, _ = fmt.Fprintln(os.Stderr, "cannot start nil server")
	}
	defer func() {
		if err := s.listener.Close(); err != nil {
			s.logger.Error("error closing listener", "error", err)
		}
	}()

	newConns := make(chan net.Conn)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				s.logger.Error("error accepting connection", "error", err)
				// @TODO: implement exponential backoff later
				time.Sleep(5 * time.Second)
				continue
			}
			newConns <- conn
		}
	}()

	for {
		select {

		case <-ctx.Done():
			// Close all connections
			s.logger.Info("gracefully shutdown server")
			_ = s.listener.Close()
			return
		case conn, ok := <-newConns:
			// NoK means newConns channel is closed.
			// Drop because we would not be able to process connections, anyway
			if !ok {
				s.logger.Debug("connections channel was closed")
			}
			go s.handleConnection(ctx, conn)
		}
	}
}

// handleConnection is the starting point of each connection established with the server.
// It reads commands from the connection, apply them and send the response back to the client.
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			s.logger.Error("error closing connection", "error", err)
		}
	}(conn)

	reader := bufio.NewReader(conn)

	for {
		select {

		case <-ctx.Done():
			s.logger.Info("initiating graceful termination", "client_ip", conn.RemoteAddr().String())
			// Done means the connection was dropped or the client is done, so immediately return
			err := conn.Close()
			if err != nil {
				s.logger.Error("error closing connection", "error", err)
			}
			return

		default:
			// Commands are sent through Array type frame.
			// Read and process them.
			cmdFrame, err := frame.Decode(reader)
			//
			if err != nil {
				if err == io.EOF {
					s.logger.Info("client initiated shutdown", "client_ip", conn.RemoteAddr().String())
					return
				}
				const errMsg = "command should be an Array frame"
				s.logger.Error(errMsg, "client_ip", conn.RemoteAddr().String(), "cmd", "Nil")
				s.SendError(errMsg, conn)
				continue
			}
			switch cmd := cmdFrame.(type) {
			case *frame.Array:
				// Process the command
				s.logger.Debug("command received", "client_ip", conn.RemoteAddr().String(), "cmd", cmd.String())
				//resp, _ := frame.NewSimpleString("PONG")
				//resp.WriteTo(conn)
			default:
				const errMsg = "command should be an Array frame"
				s.logger.Error(errMsg, "client_ip", conn.RemoteAddr().String(), "cmd", cmdFrame.String())
				s.SendError(errMsg, conn)
				continue
			}
		}
	}
}

// SendError responds to a client with an error.
// The error message should be compatible to RESP Error type (i.e. Simple String).
func (s *Server) SendError(msg string, conn net.Conn) {
	errFrame, err := frame.NewError(msg)
	if err != nil {
		s.logger.Error("error creating error frame", "error", err)
	}
	_, err = errFrame.WriteTo(conn)
	if err != nil {
		s.logger.Error("error writing error frame to network", "error", err)
	}
}
