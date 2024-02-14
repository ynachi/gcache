package main

import (
	"context"
	"fmt"
	"github.com/ynachi/gcache/server"
	"os"
	"os/signal"
	"syscall"
)

// main is for prototyping only for now.
func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	srv, err := server.NewServer("127.0.0.1", 6379, "INFO", 5000000000, "LFU")
	if err != nil {
		fmt.Printf("Error creating server %v", err)
		os.Exit(1)
	}
	fmt.Println(srv.Address())
	go srv.Start(ctx)

	// Wait for a shutdown signal
	<-sigchan
	fmt.Println("received shutdown signal")

	closed := make(chan struct{})
	go func() {
		cancel()
		// @TODO maybe add server stop method
		// srv.Stop()
		defer close(closed)
		cancel()
	}()

	// Wait for all connections to be done
	fmt.Println("waiting for connections to finish")
	<-closed
	fmt.Println("server shutdown complete")
	os.Exit(0)
}
