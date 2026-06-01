package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pennypilot/pennypilot/backend/internal/api"
	"github.com/pennypilot/pennypilot/backend/internal/config"
	"github.com/pennypilot/pennypilot/backend/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	db, err := store.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	handler := api.NewHandler(api.Dependencies{
		Config: cfg,
		Logger: logger,
		Store:  db,
	})

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      handler.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverError := make(chan error, 1)
	go func() {
		logger.Info("starting api server", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverError <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverError:
		return fmt.Errorf("api server stopped unexpectedly: %w", err)
	case sig := <-shutdown:
		logger.Info("received shutdown signal", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("api server shutdown failed: %w", err)
	}

	logger.Info("api server stopped")
	return nil
}
