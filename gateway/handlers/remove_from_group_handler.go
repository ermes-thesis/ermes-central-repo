package handlers

import (
	"ermes-central-repo/db"
	"net/http"
)

func RemoveFromGroupHandler(redisClient db.GenericDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.URL.Query().Get("userId")
		if userId == "" {
			http.Error(w, "Missing userId", http.StatusBadRequest)
			return
		}

		groupName := r.URL.Query().Get("group")
		if groupName == "" {
			http.Error(w, "Missing groupName", http.StatusBadRequest)
			return
		}

		err := redisClient.RemoveGroupFromUser(groupName, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
