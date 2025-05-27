package handlers

import (
	"centralRepoDemo/db"
	"net/http"
)

func RemoveFromGroupHandler(redisClient *db.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, httpErr := CheckHTTPReqAuth(r)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode())
		}

		groupName := r.URL.Query().Get("group")
		err := redisClient.RemoveGroupFromUser(groupName, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
