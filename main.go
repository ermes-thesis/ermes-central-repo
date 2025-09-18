package main

import (
	"context"
	"ermes-central-repo/caches"
	"ermes-central-repo/config"
	"ermes-central-repo/db"
	"ermes-central-repo/gateway"
	"ermes-central-repo/logger"
	"ermes-central-repo/tokenEncoder"
	"errors"
	"fmt"
	"github.com/ermes-thesis/ermes-wasm-pool/wasm-pool"
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
		return generalLogger.ErrorfErr("Couldn't create logger for HTTP Server component: %v\n", err)
	}

	// Create redis logger
	redisLogger, err := loggerFactory.New("redis")
	if err != nil {
		return generalLogger.ErrorfErr("Couldn't create logger Redis DB component: %v\n", err)
	}

	// Create function cache logger
	functionCacheLogger, err := loggerFactory.New("function_cache")
	if err != nil {
		return generalLogger.ErrorfErr("Couldn't create logger for Function Cache component: %v\n", err)
	}

	// Creating token encoder to release encoded group tokens
	tknEncoder, err := tokenEncoder.NewGroupTokenEncoder(cfg.JWTPemPath, cfg.JWTExpirationTime)
	if err != nil {
		generalLogger.Errorf("failed to create token encoder with JWT .pem path: %s", cfg.JWTPemPath)
		return fmt.Errorf("unable to intializate token encoder: %w", err)
	}
	generalLogger.Info("Token encoder initialized successfully")

	// Creating a WASM Pool to execute authentication function(s)
	wasmPool := wasm.NewExecutorPool(cfg.WASMPoolSize)
	generalLogger.Infof("WASM pool created successfully with %d executors", cfg.WASMPoolSize)

	// Creating a RedisClient interface to access underlying DB
	redisClient := db.NewRedisClient(cfg, redisLogger.With("component", "REDIS_DB"))
	generalLogger.Info("Redis client created successfully", "db_addr", cfg.RedisAddr, "db_port", cfg.RedisPort, "db_index", cfg.RedisDB)

	// Creating new function cache to cache retrieve authentication function(s)
	fnCache, err := caches.NewDefaultFunctionCache(cfg, functionCacheLogger.With("component", "FN_CACHE"))
	if err != nil {
		generalLogger.Error("Failed to initialize function cache", "error", err)
		return fmt.Errorf("unable to initialize function cache: %w", err)
	}
	generalLogger.Info("Function cache initialized successfully")

	// Create Group Token Provider Server Node
	generalLogger.Info("Creating a new Server instance", "wasm_pool_size", cfg.WASMPoolSize, "jwt_pem_path", cfg.JWTPemPath)
	ser := gateway.NewServer(
		cfg,
		httpLogger.With("component", "HTTP_SERVER"), //This logger will be used for HTTP Requests logging
		wasmPool,
		redisClient,
		tknEncoder,
		fnCache,
	)
	generalLogger.Info("Server initialized successfully",
		"jwt_exp_time", cfg.JWTExpirationTime,
		"http_server_port", cfg.HTTPPort,
		"jwt_pem_path", cfg.JWTPemPath,
		"wasm_pool_size", cfg.WASMPoolSize,
		"redis_addr", cfg.RedisAddr,
		"redis_port", cfg.RedisPort,
		"redis_db", cfg.RedisDB,
		"fn_cache_size", cfg.FunctionCacheSize,
		"fn_cache_bin_dir", cfg.FunctionCacheBinDir,
		"fn_cache_ttl", cfg.FunctionCacheTtl,
		"function_registry_addr", cfg.FunctionRegistryAddress,
		"function_registry_port", cfg.FunctionRegistryPort,
	)
	generalLogger.Debug("Server initialized successfully",
		"ermes_secret", cfg.ErmesSecret,
		"dev_secret", cfg.DevSecret,
		"middleware_secret", cfg.MiddlewareSecret,
		"redis_password", cfg.RedisPassword)

	// Creating HTTP Server component
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
