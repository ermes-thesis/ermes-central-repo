package handlers

import (
	"central-repo/db"
	"encoding/json"
	"net/http"
)

func CreateGroupHandler(redisClient *db.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupName := r.URL.Query().Get("group")
		if groupName == "" {
			http.Error(w, "Missing groupName", http.StatusBadRequest)
		}

		partecipantIDsJson := r.URL.Query().Get("partecipants")
		if partecipantIDsJson == "" {
			http.Error(w, "Missing partecipants", http.StatusBadRequest)
		}

		var partecipantIDs []string
		err := json.Unmarshal([]byte(partecipantIDsJson), &partecipantIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = redisClient.CreateGroup(groupName, partecipantIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
