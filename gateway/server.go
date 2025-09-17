package gateway

import (
	"ermes-central-repo/caches"
	"ermes-central-repo/config"
	"ermes-central-repo/db"
	"ermes-central-repo/logger"
	"ermes-central-repo/tokenEncoder"
	"fmt"
	"net/http"

	"github.com/ermes-thesis/ermes-wasm-pool/wasm-pool"
)

type Server struct {
	wasmPool     *wasm.ExecutorPool
	redisClient  db.GenericDB
	router       *http.ServeMux
	tokenEncoder tokenEncoder.TokenEncoder
	cache        caches.FunctionCache
	activeAuthFn string
	log          *logger.Logger
}

func (s *Server) Router() http.Handler {
	return s.router
}

func NewServer(cfg *config.Config, log *logger.Logger, dbLog *logger.Logger, fnCacheLog *logger.Logger) (*Server, error) {
	log.Info("Creating a new Server instance", "wasm_pool_size", cfg.WASMPoolSize, "jwt_pem_path", cfg.JWTPemPath)

	// Creating token encoder to release encoded group tokens
	tknEncoder, err := tokenEncoder.NewGroupTokenEncoder(cfg.JWTPemPath, cfg.JWTExpirationTime)
	if err != nil {
		log.Errorf("failed to create token encoder with JWT .pem path: %s", cfg.JWTPemPath)
		return nil, fmt.Errorf("unable to intializate token encoder: %w", err)
	}
	log.Info("Token encoder initialized successfully")

	// Creating a WASM Pool to execute authentication function(s)
	wasmPool := wasm.NewExecutorPool(cfg.WASMPoolSize)
	log.Infof("WASM pool created successfully with %d executors", cfg.WASMPoolSize)

	// Creating a RedisClient interface to access underlying DB
	redisClient := db.NewRedisClient(cfg, dbLog)
	log.Info("Redis client created successfully", "db_addr", cfg.RedisAddr, "db_port", cfg.RedisPort, "db_index", cfg.RedisDB)

	// Creating new function cache to cache retrieve authentication function(s)
	cache, err := caches.NewDefaultFunctionCache(cfg, fnCacheLog)
	if err != nil {
		log.Error("Failed to initialize function cache", "error", err)
		return nil, fmt.Errorf("unable to initialize function cache: %w", err)
	}
	log.Info("Function cache initialized successfully")

	ser := &Server{
		wasmPool:     wasmPool,
		redisClient:  redisClient,
		log:          log,
		tokenEncoder: tknEncoder,
		cache:        cache,
		router:       http.NewServeMux(),
	}
	ser.initRouter(cfg)

	log.Info("Server initialized successfully",
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
	return ser, nil
}

func (s *Server) Close() error {
	s.log.Info("Closing the server...")
	err := s.redisClient.Close()
	if err != nil {
		s.log.Error("Failed to close Redis client", "error", err)
		return fmt.Errorf("could not close Redis client: %w", err)
	}
	s.log.Info("Redis client closed successfully")

	s.wasmPool.Close()
	s.log.Info("WASM pool closed successfully")
	s.log.Info("Server closed successfully")
	return nil
}

func (s *Server) Init(initFun func(client db.GenericDB)) {
	initFun(s.redisClient)
	//err := s.redisClient.CreateGroup("DOCTORS", []string{})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = s.redisClient.AddGroupToUser("DOCTORS", "1")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//groups, err := s.redisClient.GetUserGroups("1")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("The groups for user:%d, are: %v\n", 1, groups)
	//
	//err = s.redisClient.CreateGroup("PROVA", []string{})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//existingGroups, err := s.redisClient.GetExistingGroups()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("The existing groups are: %v\n", existingGroups)
	//err = s.redisClient.DeleteGroup("PROVA")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//existingGroups, err = s.redisClient.GetExistingGroups()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("The existing groups are: %v\n", existingGroups)
}
