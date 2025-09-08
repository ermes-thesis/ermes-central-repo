package gateway

import (
	"ermes-central-repo/config"
	"ermes-central-repo/gateway/handlers"
	"io"
	"net/http"
)

func (s *Server) initRouter() {
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{ "message": "Hello World" }`)
	})
	s.router.HandleFunc("GET /token", filterMiddlewareSecret(handlers.RetrieveGroupToken(s.redisClient, s.wasmPool, s.tokenEncoder, &s.activeAuthFn)))

	s.router.HandleFunc("POST /group/{groupName}/{userId}", filterErmesSecret(handlers.AddToGroupHandler(s.redisClient)))

	s.router.HandleFunc("DELETE /group/{groupName}/{userId}", filterErmesSecret(handlers.RemoveFromGroupHandler(s.redisClient)))

	s.router.HandleFunc("POST /group/{groupName}", filterErmesSecret(handlers.CreateGroupHandler(s.redisClient)))

	s.router.HandleFunc("DELETE /group/{groupName}", filterErmesSecret(handlers.DeleteGroupHandler(s.redisClient)))

	s.router.HandleFunc("PUT /set-auth-function/{name}/{version}", filterDevSecret(handlers.SetAuthFunctionHandler(s.cache, &s.activeAuthFn)))
	s.router.HandleFunc("PUT /set-auth-function/{name}", filterDevSecret(handlers.SetAuthFunctionHandler(s.cache, &s.activeAuthFn)))
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

func filterDevSecret(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-DEV-SECRET") != config.DevSecret() {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}
