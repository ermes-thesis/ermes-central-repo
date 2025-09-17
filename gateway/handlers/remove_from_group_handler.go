package handlers

import (
	"ermes-central-repo/db"
	"ermes-central-repo/logger"
	"fmt"
	"net/http"
)

func RemoveFromGroupHandler(redisClient db.GenericDB, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		userId := r.PathValue("userId")
		if userId == "" {
			log.Warn("Missing userId in request path",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			log.Debug("Missing userId in request path", "headers", r.Header)
			http.Error(w, "Missing userId in request path", http.StatusBadRequest)
			return
		}

		groupName := r.PathValue("groupName")
		if groupName == "" {
			log.Warn("Missing groupName in request path",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			log.Debug("Missing groupName in request path",
				"user_id", userId,
				"headers", r.Header,
			)
			http.Error(w, "Missing groupName in request path", http.StatusBadRequest)
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
