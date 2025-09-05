package gateway

import (
	"ermes-central-repo/config"
	"ermes-central-repo/db"
	"ermes-central-repo/logger"
	"ermes-central-repo/tokenEncoder"
	"github.com/ermes-thesis/ermes-wasm-pool/wasm-pool"
	"net/http"
)

type Server struct {
	wasmPool    *wasm.ExecutorPool
	redisClient db.GenericDB
	router      http.Handler
	logger      *logger.Logger
}

func (s *Server) Router() http.Handler {
	return s.router
}

func NewServer(logger *logger.Logger) (*Server, error) {
	logger.Info("Creating a new Server instance... \n")

	tknEncoder, err := tokenEncoder.NewGroupTokenEncoder(config.JWTPemPath())
	if err != nil {
		return nil, err
	}
	wasmPool := wasm.NewExecutorPool(config.WASMPoolSize())
	redisClient := db.NewRedisClient()
	ser := &Server{
		wasmPool:    wasmPool,
		redisClient: redisClient,
		logger:      logger,
		router:      NewRouter(redisClient, wasmPool, tknEncoder),
	}
	logger.Info("done\n")
	return ser, nil
}

func (s *Server) Close() error {
	err := s.redisClient.Close()
	if err != nil {
		return err
	}

	s.wasmPool.Close()
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
