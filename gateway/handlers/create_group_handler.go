package handlers

import (
	"encoding/json"
	"ermes-central-repo/db"
	"net/http"
)

func CreateGroupHandler(redisClient db.GenericDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupName := r.PathValue("groupName")
		if groupName == "" {
			http.Error(w, "Missing groupName", http.StatusBadRequest)
			return
		}

		var participantIDs []string
		err := json.NewDecoder(r.Body).Decode(&participantIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = r.Body.Close()

		err = redisClient.CreateGroup(groupName, participantIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
