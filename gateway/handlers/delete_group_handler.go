package handlers

import (
	"ermes-central-repo/db"
	"net/http"
)

func DeleteGroupHandler(redisClient db.GenericDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.Body.Close()
		groupName := r.PathValue("groupName")
		if groupName == "" {
			http.Error(w, "Missing groupName", http.StatusBadRequest)
			return
		}

		err := redisClient.DeleteGroup(groupName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
