package handlers

import (
	"encoding/json"
	"ermes-central-repo/db"
	"net/http"
)

func CreateGroupHandler(redisClient db.GenericDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupName := r.URL.Query().Get("group")
		if groupName == "" {
			http.Error(w, "Missing groupName", http.StatusBadRequest)
		}

		participantIDsJson := r.URL.Query().Get("participants")
		if participantIDsJson == "" {
			http.Error(w, "Missing participants", http.StatusBadRequest)
		}

		var participantIDs []string
		err := json.Unmarshal([]byte(participantIDsJson), &participantIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = redisClient.CreateGroup(groupName, participantIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
