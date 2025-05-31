package gateway

import (
	"centralRepoDemo/db"
	"centralRepoDemo/gateway/handlers"
	"github.com/ermes-thesis/ermes-wasm-test/wasm-pool"
	"io"
	"net/http"
	"time"
)

func NewRouter(redisClient *db.RedisClient, wasmPool *wasm.ExecutorPool) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{ "message": "Hello World" }`)
		time.Sleep(20 * time.Second)
	})
	mux.HandleFunc("/group-access", handlers.RetrieveGroupToken(redisClient, wasmPool))

	mux.HandleFunc("/add-to-group", handlers.AddToGroupHandler(redisClient, wasmPool))

	mux.HandleFunc("/remove-from-group", handlers.RemoveFromGroupHandler(redisClient, wasmPool))

	mux.HandleFunc("/create-group", handlers.CreateGroupHandler(redisClient, wasmPool))

	mux.HandleFunc("/delete-group", handlers.DeleteGroupHandler(redisClient, wasmPool))

	return mux
}
