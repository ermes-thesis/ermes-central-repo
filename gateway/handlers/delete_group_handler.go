package handlers

import (
	"central-repo/db"
	"net/http"
)

func DeleteGroupHandler(redisClient *db.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupName := r.URL.Query().Get("group")
		if groupName == "" {
			http.Error(w, "Missing groupName", http.StatusBadRequest)
		}
		err := redisClient.DeleteGroup(groupName)

		//TODO: Modify the code in order to also remove users belonging to the 'groupName' group

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
