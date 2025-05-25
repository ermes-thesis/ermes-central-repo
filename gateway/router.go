package gateway

import (
	"centralRepoDemo/db/postgres/DAO"
	"centralRepoDemo/db/redis"
	"centralRepoDemo/gateway/handlers"
	"io"
	"net/http"
	"time"
)

func NewRouterWithPostgres(daoFactory *DAO.DAOFactory) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{ "message": "Hello World" }`)
		time.Sleep(20 * time.Second)
	})
	mux.HandleFunc("/group-access", handlers.RetrieveGroupToken(daoFactory))

	return mux
}

func NewRouterWithRedis(redisClient *redis.RedisClient) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{ "message": "Hello World" }`)
		time.Sleep(20 * time.Second)
	})
	mux.HandleFunc("/group-access", handlers.RetrieveGroupTokenWithRedis(redisClient))

	mux.HandleFunc("/add-to-group", handlers.AddToGroupHandler(redisClient))

	mux.HandleFunc("/remove-from-group", handlers.RemoveFromGroupHandler(redisClient))

	mux.HandleFunc("/create-group", handlers.CreateGroupHandler(redisClient))

	mux.HandleFunc("/delete-group", handlers.DeleteGroupHandler(redisClient))

	return mux
}
