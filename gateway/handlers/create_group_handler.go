package handlers

import (
	"encoding/json"
	"ermes-central-repo/db"
	"fmt"
	"net/http"

	"github.com/ermes-thesis/ermes-common/logger"
)

func CreateGroupHandler(redisClient db.GenericDB, log logger.Logger) http.HandlerFunc {
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

		var participantIDs []string

		if r.Body != http.NoBody {
			if err := json.NewDecoder(r.Body).Decode(&participantIDs); err != nil {
				log.Warn("Failed to decode participant IDs",
					"remote_addr", r.RemoteAddr,
					"method", r.Method,
					"path", r.URL.RequestURI(),
					"error", err,
				)
				log.Debug("Failed to decode participant IDs",
					"group_name", groupName)
				http.Error(w, fmt.Sprintf("Invalid participant list: %v", err), http.StatusBadRequest)
				return
			}
		}

		if err := redisClient.CreateGroup(groupName, participantIDs); err != nil {
			log.Warn("Failed to create group",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"error", err,
			)
			log.Debug("Failed to create group",
				"group_name", groupName,
				"participant_ids", participantIDs,
			)
			http.Error(w, fmt.Sprintf("Could not create group '%s': %v", groupName, err), http.StatusInternalServerError)
			return
		}

		log.InfofWith("Group '%s' created successfully",
			[]any{groupName},
			"remote_addr", r.RemoteAddr,
			"method", r.Method,
			"path", r.URL.RequestURI(),
		)
		w.WriteHeader(http.StatusCreated)
	}
}
