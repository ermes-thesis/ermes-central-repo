package main

import (
	"context"
	"ermes-central-repo/caches"
	"ermes-central-repo/config"
	"ermes-central-repo/db"
	"ermes-central-repo/gateway"
	"ermes-central-repo/tokenEncoder"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/ermes-thesis/ermes-common/wasm-pool"
	"github.com/spf13/pflag"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while running Group Token Provider: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var cfgPath string
	pflag.StringVarP(&cfgPath, "config", "c", "", "path to config file")
	pflag.Parse()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Create logger factory

	loggerFactory := logger.NewFactory(logger.Config{
		Level:  logger.InfoLevel,
		Format: "text",
		Output: &logger.LogFile{
			LogDir: "gtpLogs",
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
	cfg, err := config.LoadConfig(generalLogger, cfgPath)
	if err != nil {
		return generalLogger.ErrorfErr("error loading configuration: %v", err)
	}
	generalLogger.Infof("Configuration loaded successfully: %v", cfg)

	// Creating token encoder to release encoded group tokens
	tknEncoder, err := tokenEncoder.NewGroupTokenEncoder(cfg.JWTPemPath, cfg.JWTExpirationTime)
	if err != nil {
		generalLogger.Errorf("failed to create token encoder with JWT .pem path: %s", cfg.JWTPemPath)
		return fmt.Errorf("unable to intializate token encoder: %w", err)
	}
	generalLogger.Info("Token encoder initialized successfully")

	// Creating a WASM Pool to execute authentication function(s)
	var wasmPool wasm.ExecutorPool
	switch cfg.WasmRuntime {
	case wasm.Wasmedge:
		generalLogger.Info("WASM runtime selected: Wasmedge")
		wasmPool = wasm.NewWasmedgeExecutorPool(cfg.WasmPoolConfig, generalLogger.Component("WASM_POOL"))
	case wasm.Wasmtime:
		generalLogger.Info("WASM runtime selected: Wasmtime")
		wasmPool = wasm.NewWasmtimeExecutorPool(cfg.WasmPoolConfig, generalLogger.Component("WASM_POOL"))
	case wasm.Wazero:
		generalLogger.Info("WASM runtime selected: Wazero")
		wasmPool = wasm.NewWazeroExecutorPool(cfg.WasmPoolConfig, ctx, generalLogger.Component("WASM_POOL"))
	default:
		generalLogger.Errorf("invalid WASM runtime selected: %s", cfg.WasmRuntime)
		return fmt.Errorf("invalid WASM runtime selected: %s", cfg.WasmRuntime)
	}
	generalLogger.Infof("WASM pool created successfully with %d executors", cfg.WasmPoolConfig.PoolSize)

	// Creating a RedisClient interface to access underlying DB
	redisClient, err := db.NewRedisClient(cfg, generalLogger.Component("REDIS_DB"))
	if err != nil {
		generalLogger.Error("Failed to initialize redis client", "error", err)
		return fmt.Errorf("unable to initialize redis client: %w", err)
	}
	generalLogger.Info("Redis client created successfully", "db_addr", cfg.RedisAddr, "db_port", cfg.RedisPort, "db_index", cfg.RedisDB)

	// Creating new function cache to cache retrieve authentication function(s)
	fnCache, err := caches.NewDefaultFunctionCache(cfg, generalLogger.Component("FN_CACHE"))
	if err != nil {
		generalLogger.Error("Failed to initialize function cache", "error", err)
		return fmt.Errorf("unable to initialize function cache: %w", err)
	}
	generalLogger.Info("Function cache initialized successfully")

	// Create Group Token Provider Server Node
	generalLogger.Info("Creating a new Server instance", "wasm_pool_size", cfg.WasmPoolConfig.PoolSize, "jwt_pem_path", cfg.JWTPemPath)
	ser := gateway.NewServer(
		cfg,
		generalLogger.Component("HTTP_SERVER"), //This logger will be used for HTTP Requests logging
		wasmPool,
		redisClient,
		tknEncoder,
		fnCache,
	)
	generalLogger.Info("Server initialized successfully",
		"jwt_exp_time", cfg.JWTExpirationTime,
		"http_server_port", cfg.HTTPPort,
		"jwt_pem_path", cfg.JWTPemPath,
		"wasm_pool_size", cfg.WasmPoolConfig.PoolSize,
		"redis_addr", cfg.RedisAddr,
		"redis_port", cfg.RedisPort,
		"redis_db", cfg.RedisDB,
		"fn_cache_size", cfg.FunctionCache.Size,
		"fn_cache_bin_dir", cfg.FunctionCache.BinDir,
		"fn_cache_ttl", cfg.FunctionCache.Ttl,
		"function_registry_addr", cfg.FunctionRegistryAddress,
		"function_registry_port", cfg.FunctionRegistryPort,
	)
	generalLogger.Debug("Server initialized successfully",
		"ermes_secret", cfg.ErmesSecret,
		"dev_secret", cfg.DevSecret,
		"client_secret", cfg.ClientSecret,
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
