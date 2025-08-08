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
	log, err := logger.NewLogger(logger.Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: &logger.LogFile{
			LogDir:      "logs",
			LogFileName: "central_repo.log",
		},
		AddSource:     false,
		StdoutEnabled: true,
	})
	defer log.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := run(ctx, log); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	//Run server
	ser := gateway.NewServer(log)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort("", config.HTTPPort()),
		Handler: ser.Router(),
	}

	go func() {
		log.Info("Server correctly created and listening on port", "http_port", config.HTTPPort())

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
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
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
		if err := ser.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
		log.Info("http server shut down successfully\n")
	}()
	wg.Wait()
	return nil
}
