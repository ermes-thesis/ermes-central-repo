package handlers

import (
	"ermes-central-repo/caches"
	"fmt"
	"net/http"

	"github.com/ermes-thesis/ermes-common/logger"
)

func SetAuthFunctionHandler(cache caches.FunctionCache, activeAuthFun *string, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		fnName := r.PathValue("name")
		if fnName == "" {
			log.Warn("Missing function name in path request",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			http.Error(w, "Missing function name in path", http.StatusBadRequest)
			return
		}
		fnVersion := r.PathValue("version")
		if fnVersion == "" {
			fnVersion = "latest"
		}

		metadata, err := cache.GetMetadata(fnName, fnVersion)
		if err != nil {
			log.Warn("Failed to retrieve function metadata",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"error", err,
			)
			log.Debug("Failed to retrieve function metadata",
				"fn_name", fnName,
				"fn_version", fnVersion,
				"headers", r.Header,
			)
			http.Error(w, fmt.Sprintf("Function '%s@%s' not found", fnName, fnName), http.StatusNotFound)
			return
		}

		*activeAuthFun = metadata.Name + "@" + metadata.Version
		log.InfofWith("Authentication function set successfully to '%s'",
			[]any{*activeAuthFun},
			"remote_addr", r.RemoteAddr,
			"method", r.Method,
			"path", r.URL.RequestURI(),
		)

		w.WriteHeader(http.StatusOK)
	}
}
