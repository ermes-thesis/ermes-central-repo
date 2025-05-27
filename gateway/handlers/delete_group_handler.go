package handlers

import (
	"centralRepoDemo/db"
	"net/http"
)

func DeleteGroupHandler(redisClient *db.RedisClient) http.HandlerFunc {
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
