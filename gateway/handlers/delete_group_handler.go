package handlers

import (
	"ermes-central-repo/db"
	"fmt"
	"net/http"

	"github.com/ermes-thesis/ermes-common/logger"
)

func DeleteGroupHandler(redisClient db.GenericDB, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		groupName := r.PathValue("groupName")
		if groupName == "" {
			log.Warn("Missing groupName in request path",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			http.Error(w, "Missing groupName in path", http.StatusBadRequest)
			return
		}

		if err := redisClient.DeleteGroup(groupName); err != nil {
			log.Warn("Failed to delete group",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"error", err,
			)
			log.Debug("Failed to delete group",
				"group_name", groupName)
			http.Error(w, fmt.Sprintf("Could not delete group '%s': %v", groupName, err), http.StatusInternalServerError)
			return
		}

		log.InfofWith("Deleted group '%s' successfully",
			[]any{groupName},
			"remote_addr", r.RemoteAddr,
			"method", r.Method,
			"path", r.URL.RequestURI(),
		)
		w.WriteHeader(http.StatusOK)
	}
}
