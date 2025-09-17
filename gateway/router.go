package gateway

import (
	"ermes-central-repo/gateway/handlers"
	"ermes-central-repo/logger"
	"io"
	"net/http"
)

func (s *Server) initRouter() {
	ermesClientLibSecret := s.nodeCfg.MiddlewareSecret
	ermesSecret := s.nodeCfg.ErmesSecret
	devSecret := s.nodeCfg.DevSecret

	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{ "message": "Hello World" }`)
	})
	s.router.HandleFunc("GET /token", filterMiddlewareSecret(s.log, ermesClientLibSecret, handlers.RetrieveGroupToken(s.redisClient, s.wasmPool, s.tokenEncoder, &s.activeAuthFn, s.log, s.nodeCfg.FunctionCacheBinDir)))

	s.router.HandleFunc("POST /group/{groupName}/{userId}", filterErmesSecret(s.log, ermesSecret, handlers.AddToGroupHandler(s.redisClient, s.log)))

	s.router.HandleFunc("DELETE /group/{groupName}/{userId}", filterErmesSecret(s.log, ermesSecret, handlers.RemoveFromGroupHandler(s.redisClient, s.log)))

	s.router.HandleFunc("POST /group/{groupName}", filterErmesSecret(s.log, ermesSecret, handlers.CreateGroupHandler(s.redisClient, s.log)))

	s.router.HandleFunc("DELETE /group/{groupName}", filterErmesSecret(s.log, ermesSecret, handlers.DeleteGroupHandler(s.redisClient, s.log)))

	s.router.HandleFunc("PUT /set-auth-function/{name}/{version}", filterDevSecret(s.log, devSecret, handlers.SetAuthFunctionHandler(s.cache, &s.activeAuthFn, s.log)))
	s.router.HandleFunc("PUT /set-auth-function/{name}", filterDevSecret(s.log, devSecret, handlers.SetAuthFunctionHandler(s.cache, &s.activeAuthFn, s.log)))
}

func filterErmesSecret(log *logger.Logger, ermesSecret string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-ERMES-SECRET") != ermesSecret {
			log.Warn("Unauthorized request with invalid Ermes secret",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			http.Error(w, "Forbidden: invalid Ermes secret", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func filterMiddlewareSecret(log *logger.Logger, ermesClientLibrarySecret string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-MIDDLEWARE-SECRET") != ermesClientLibrarySecret {
			log.Warn("Unauthorized request with invalid Ermes Client library secret",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			http.Error(w, "Forbidden: invalid Ermes client library secret", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func filterDevSecret(log *logger.Logger, devSecret string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-DEV-SECRET") != devSecret {
			log.Warn("Unauthorized request with invalid Ermes Dev Secret",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			http.Error(w, "Forbidden: invalid Dev secret", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}
