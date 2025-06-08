package gateway

import (
	"ermes-central-repo/db"
	"ermes-central-repo/gateway/handlers"
	"github.com/ermes-thesis/ermes-wasm-pool/wasm-pool"
	"io"
	"net/http"
)

func NewRouter(redisClient db.GenericDB, wasmPool *wasm.ExecutorPool) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{ "message": "Hello World" }`)
	})
	mux.HandleFunc("GET /token", handlers.RetrieveGroupToken(redisClient, wasmPool))

	mux.HandleFunc("POST /group/{groupName}/{userId}", handlers.AddToGroupHandler(redisClient))

	mux.HandleFunc("DELETE /group/{groupName}/{userId}", handlers.RemoveFromGroupHandler(redisClient))

	mux.HandleFunc("POST /group/{groupName}", handlers.CreateGroupHandler(redisClient))

	mux.HandleFunc("DELETE /group/{groupName}", handlers.DeleteGroupHandler(redisClient))

	return mux
}
