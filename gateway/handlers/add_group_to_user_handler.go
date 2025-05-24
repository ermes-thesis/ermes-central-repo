package handlers

import (
	"centralRepoDemo/db/redis"
	"net/http"
	"strings"
)

func AddToGroupHandler(redisClient *redis.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No authorization header", http.StatusBadRequest)
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			http.Error(w, "Invalid authorization header format", http.StatusBadRequest)
			return
		}

		//userId, err := executeAuthFunction()
		//if err != nil {
		//	http.Error(w, "Authentication unsuccessful", http.StatusUnauthorized)
		//}
		userId := 2
		groupName := r.URL.Query().Get("group")

		err := redisClient.AddGroupToUser(groupName, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
