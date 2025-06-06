package handlers

import (
	"ermes-central-repo/db"
	"net/http"
)

func DeleteGroupHandler(redisClient db.GenericDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupName := r.URL.Query().Get("group")
		if groupName == "" {
			http.Error(w, "Missing groupName", http.StatusBadRequest)
		}
		err := redisClient.DeleteGroup(groupName)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
