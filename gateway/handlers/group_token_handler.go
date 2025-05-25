package handlers

import (
	"centralRepoDemo/db/postgres/DAO"
	"centralRepoDemo/db/redis"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func RetrieveGroupToken(factory *DAO.DAOFactory) http.HandlerFunc {
	usersGroupsDAO := factory.UserGroupRelation()
	log.Printf("Created usersGroupsDAO to access users_groups table: %+v\n", usersGroupsDAO)

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No authorization header", http.StatusBadRequest)
			return
		}

		if !strings.HasPrefix(authHeader, AUTH_PREFIX) {
			http.Error(w, "Invalid authorization header format", http.StatusBadRequest)
		}

		userId, err := executeAuthFunction()
		if err != nil {
			http.Error(w, "Authentication unsuccessful", http.StatusUnauthorized)
		}

		res, err := usersGroupsDAO.GetByUserId(userId)

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

func RetrieveGroupTokenWithRedis(rdbClient *redis.RedisClient) http.HandlerFunc {
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
