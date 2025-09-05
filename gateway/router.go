package gateway

import (
	"ermes-central-repo/config"
	"ermes-central-repo/db"
	"ermes-central-repo/gateway/handlers"
	"ermes-central-repo/tokenEncoder"
	"github.com/ermes-thesis/ermes-wasm-pool/wasm-pool"
	"io"
	"net/http"
)

func NewRouter(redisClient db.GenericDB, wasmPool *wasm.ExecutorPool, tknEncoder tokenEncoder.TokenEncoder) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{ "message": "Hello World" }`)
	})
	mux.HandleFunc("GET /token", filterMiddlewareSecret(handlers.RetrieveGroupToken(redisClient, wasmPool, tknEncoder)))

	mux.HandleFunc("POST /group/{groupName}/{userId}", filterErmesSecret(handlers.AddToGroupHandler(redisClient)))

	mux.HandleFunc("DELETE /group/{groupName}/{userId}", filterErmesSecret(handlers.RemoveFromGroupHandler(redisClient)))

	mux.HandleFunc("POST /group/{groupName}", filterErmesSecret(handlers.CreateGroupHandler(redisClient)))

	mux.HandleFunc("DELETE /group/{groupName}", filterErmesSecret(handlers.DeleteGroupHandler(redisClient)))

	return mux
}

func filterErmesSecret(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-ERMES-SECRET") != config.ErmesSecret() {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func filterMiddlewareSecret(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-MIDDLEWARE-SECRET") != config.MiddlewareSecret() {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}
