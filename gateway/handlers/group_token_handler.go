package handlers

import (
	"centralRepoDemo/db/redis"
	"encoding/json"
	"fmt"
	"net/http"
)

func RetrieveGroupToken(rdbClient *redis.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, httpErr := CheckHTTPReqAuth(r)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode())
			return
		}

		res, err := rdbClient.GetUserGroups(userId)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error while retrieving access group %s", err), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			return
		}
	}

}
