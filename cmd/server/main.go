package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fizz-buzz/internal/config"
	"fizz-buzz/internal/httpapi"
	"fizz-buzz/internal/stats"
)

const (
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, warnings := config.LoadFromEnv()

	for _, warning := range warnings {
		logger.Warn("configuration fallback applied", "field", warning.Field, "message", warning.Message, "value", warning.Value)
	}

	statsStore, err := stats.NewSQLiteStoreWithTimeout(cfg.StatsDBDSN, cfg.StatsDBTimeout)
	if err != nil {
		logger.Error("failed to initialize stats store", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := statsStore.Close(); closeErr != nil {
			logger.Error("failed to close stats store", "error", closeErr)
		}
	}()

	handler := httpapi.NewHandler(statsStore, cfg.MaxLimit)

	// Configure conservative defaults so the server behaves safely in production-like environments.
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		logger.Info("server starting", "addr", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	// Give in-flight requests a bounded amount of time to complete before exiting.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
