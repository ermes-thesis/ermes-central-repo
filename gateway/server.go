package gateway

import (
	"ermes-central-repo/caches"
	"ermes-central-repo/config"
	"ermes-central-repo/db"
	"ermes-central-repo/tokenEncoder"
	"fmt"
	"net/http"

	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/ermes-thesis/ermes-common/wasm-pool"
)

type Server struct {
	nodeCfg      *config.Config
	wasmPool     wasm.WASP
	redisClient  db.GenericDB
	router       *http.ServeMux
	tokenEncoder tokenEncoder.TokenEncoder
	cache        caches.FunctionCache
	activeAuthFn string
	log          logger.Logger
}

func (s *Server) Router() http.Handler {
	return s.router
}

func NewServer(cfg *config.Config, log logger.Logger, wasmPool wasm.WASP, redis db.GenericDB, tknEncoder tokenEncoder.TokenEncoder, fnCache caches.FunctionCache) *Server {
	ser := &Server{
		nodeCfg:      cfg,
		wasmPool:     wasmPool,
		redisClient:  redis,
		log:          log,
		tokenEncoder: tknEncoder,
		cache:        fnCache,
		router:       http.NewServeMux(),
		activeAuthFn: cfg.DefaultAuthFunction,
	}
	ser.initRouter()

	return ser
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
