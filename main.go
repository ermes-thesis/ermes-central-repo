package main

import (
	"context"
	"ermes-central-repo/config"
	"ermes-central-repo/gateway"
	"ermes-central-repo/logger"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while running Group Token Provider: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Create logger factory
	loggerFactory := logger.NewFactory(logger.Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &logger.LogFile{
			LogDir: "logs",
		},
		StdoutEnabled: true,
		AddSource:     true,
	})

	// Create general logger
	generalLogger, err := loggerFactory.New("general")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Couldn't create logger: %v\n", err)
		os.Exit(1)
	}
	defer generalLogger.Close()

	// Load configuration
	cfg, err := config.LoadConfig("config.yaml", generalLogger)
	if err != nil {
		return generalLogger.ErrorfErr("error loading configuration: %v", err)
	}
	generalLogger.Infof("Configuration loaded successfully: %v", cfg)

	// Create HTTP Server logger
	httpLogger, err := loggerFactory.New("http_server")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Couldn't create logger for HTTP Server component: %v\n", err)
		os.Exit(1)
	}

	// Create redis logger
	redisLogger, err := loggerFactory.New("redis")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Couldn't create logger Redis DB component: %v\n", err)
		os.Exit(1)
	}

	// Create function cache logger
	functionCacheLogger, err := loggerFactory.New("function_cache")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Couldn't create logger for Function Cache component: %v\n", err)
		os.Exit(1)
	}

	//Run server
	ser, err := gateway.NewServer(cfg, httpLogger.With("component", "HTTP_SERVER"), redisLogger.With("component", "REDIS_DB"), functionCacheLogger.With("component", "FN_CACHE"))
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:    net.JoinHostPort("", cfg.HTTPPort),
		Handler: ser.Router(),
	}

	go func() {
		generalLogger.Info("Server correctly created and listening on port", "http_port", cfg.HTTPPort)

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			generalLogger.Errorf("error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			generalLogger.Errorf("error shutting down http server: %s\n", err)
		}
		if err := ser.Close(); err != nil {
			generalLogger.Errorf("error shutting down http server: %s\n", err)
		}
		generalLogger.Info("http server shut down successfully\n")
	}()
	wg.Wait()
	return nil
}
