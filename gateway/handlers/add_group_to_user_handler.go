package handlers

import (
	"ermes-central-repo/db"
	"fmt"
	"net/http"

	"github.com/ermes-thesis/ermes-common/logger"
)

func AddToGroupHandler(redisClient db.GenericDB, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		groupName := r.PathValue("groupName")
		if groupName == "" {
			log.Warn("Missing groupName in path request",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			http.Error(w, "Missing groupName in path", http.StatusBadRequest)
			return
		}

		userId := r.PathValue("userId")
		if userId == "" {
			log.Warn("Missing userId in path request",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
			)
			log.Debug("Missing userId in path request",
				"group_name", groupName,
			)
			http.Error(w, "Missing userId in path", http.StatusBadRequest)
			return
		}

		err := redisClient.AddGroupToUser(groupName, userId)
		if err != nil {
			log.Warn("Failed to add user to group",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"error", err,
			)
			log.Debug("Failed to add user to group",
				"user_id", userId,
				"group_name", groupName,
			)
			http.Error(w, fmt.Sprintf("Internal Server Error: could not add user %s to group %s: %v", userId, groupName, err), http.StatusInternalServerError)
			return
		}

		log.InfofWith("User successfully added to group: %s -> %s",
			[]any{userId, groupName},
			"remote_addr", r.RemoteAddr,
			"method", r.Method,
			"path", r.URL.RequestURI(),
		)

		w.WriteHeader(http.StatusOK)
	}
}
