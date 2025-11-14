package handlers

import (
	"ermes-central-repo/db"
	"fmt"
	"net/http"

	"github.com/ermes-thesis/ermes-common/logger"
)

func RemoveFromGroupHandler(redisClient db.GenericDB, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		groupName := r.PathValue("groupName")
		if groupName == "" {
			log.Warn("Missing groupName in request path",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			log.Debug("Missing groupName in request path",
				"headers", r.Header,
			)
			http.Error(w, "Missing groupName in path", http.StatusBadRequest)
			return
		}

		userId := r.PathValue("userId")
		if userId == "" {
			log.Warn("Missing userId in request path",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			log.Debug("Missing userId in request path", "headers", r.Header, "user_id", userId)
			http.Error(w, "Missing userId in path", http.StatusBadRequest)
			return
		}

		if err := redisClient.RemoveGroupFromUser(groupName, userId); err != nil {
			log.Warn("Failed to remove from group",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"error", err,
			)
			log.Debug("Failed to remove from group",
				"user_id", userId,
				"group_name", groupName,
				"headers", r.Header,
			)
			http.Error(w, fmt.Sprintf("Could not remove user '%s' from group '%s': %v", userId, groupName, err), http.StatusInternalServerError)
			return
		}

		log.InfofWith("User '%s' successfully removed from group '%s'.",
			[]any{userId, groupName},
			"remote_addr", r.RemoteAddr,
			"method", r.Method,
			"path", r.URL.RequestURI(),
		)
		w.WriteHeader(http.StatusOK)
	}
}
