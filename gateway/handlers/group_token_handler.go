package handlers

import (
	"centralRepoDemo/config"
	"centralRepoDemo/db"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

func RetrieveGroupToken(rdbClient *db.RedisClient) http.HandlerFunc {
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

		groupToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"issuer":      userId,
			"permissions": res,
			"exp":         time.Now().Add(time.Hour * 1).Unix(), // expires in 1 hour
		})

		groupTokenString, err := groupToken.SignedString([]byte(config.JWTSecret()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"group_token": groupTokenString,
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			return
		}
	}

}
