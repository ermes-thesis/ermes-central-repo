package gateway

import (
	"ermes-central-repo/config"
	"ermes-central-repo/db"
	"fmt"
	"github.com/ermes-thesis/ermes-wasm-pool/wasm-pool"
	"io"
	"net/http"
)

type Server struct {
	wasmPool    *wasm.ExecutorPool
	redisClient db.GenericDB
	router      http.Handler
	logger      io.Writer
}

func (s *Server) Router() http.Handler {
	return s.router
}

func NewServer(logger io.Writer) *Server {
	_, _ = fmt.Fprintf(logger, "Creating a new Server instance... \n")

	wasmPool := wasm.NewExecutorPool(config.WASMPoolSize())
	redisClient := db.NewRedisClient()
	ser := &Server{
		wasmPool:    wasmPool,
		redisClient: redisClient,
		logger:      logger,
		router:      filterMiddleware(NewRouter(redisClient, wasmPool)),
	}
	_, _ = fmt.Fprintf(logger, "done \n")
	return ser
}

func filterMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-ERMES-SECRET") != config.ErmesSecret() {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
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
