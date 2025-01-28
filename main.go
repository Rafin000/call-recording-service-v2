package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/server"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeouts.Server.Startup)
	defer cancel()

	// Create a new server instance
	s, err := server.NewServer(ctx)
	if err != nil {
		slog.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	errCh := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		errCh <- s.Start()
	}()

	// Graceful shutdown handling
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		// If an error occurs while starting the server, log it
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
		}
	case <-shutdownCh:
		// If a shutdown signal is received, gracefully shut down the server
		ctx, cancel := context.WithTimeout(context.Background(), common.Timeouts.Server.Write)
		defer cancel()

		// Shutdown the server and handle any errors
		if err := s.Shutdown(ctx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
		}
	}

	// Log server stop message
	slog.Info("Server stopped gracefully")
}
