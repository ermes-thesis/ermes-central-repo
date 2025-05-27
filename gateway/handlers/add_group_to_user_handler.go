package handlers

import (
	"centralRepoDemo/db"
	"net/http"
)

func AddToGroupHandler(redisClient *db.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, httpErr := CheckHTTPReqAuth(r)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode())
		}

		groupName := r.URL.Query().Get("group")

		err := redisClient.AddGroupToUser(groupName, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
