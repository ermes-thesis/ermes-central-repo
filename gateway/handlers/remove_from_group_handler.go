package handlers

import (
	"centralRepoDemo/db/redis"
	"net/http"
)

func RemoveFromGroupHandler(redisClient *redis.RedisClient) http.HandlerFunc {
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
