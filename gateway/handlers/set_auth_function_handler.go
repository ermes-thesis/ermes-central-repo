package handlers

import (
	"caches"
	"net/http"
)

func SetAuthFunctionHandler(cache *caches.MetadataBinaryCache, activeAuthFun *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.Body.Close()

		fnName := r.PathValue("name")
		fnVersion := r.PathValue("version")

		if fnVersion == "" {
			fnVersion = "latest"
		}

		metadata, err := cache.Get(fnName, fnVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		*activeAuthFun = metadata.Name + "@" + metadata.Version

		w.WriteHeader(http.StatusOK)
	}
}
