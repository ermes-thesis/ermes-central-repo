package handlers

import (
	"centralRepoDemo/db/redis"
	"net/http"
)

func DeleteGroupHandler(redisClient *redis.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, httpErr := CheckHTTPReqAuth(r)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode())
		}

		groupName := r.URL.Query().Get("group")
		err := redisClient.DeleteGroup(groupName)

		//TODO: Modify the code in order to also remove users belonging to the 'groupName' group

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
