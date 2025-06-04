package handlers

import (
	"encoding/json"
	"ermes-central-repo/config"
	"ermes-central-repo/db"
	"fmt"
	"github.com/ermes-thesis/ermes-wasm-pool/wasm-pool"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

func RetrieveGroupToken(rdbClient db.GenericDB, wasmPool *wasm.ExecutorPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, httpErr := checkHTTPReqAuth(r, wasmPool)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode())
			return
		}
		res, err := rdbClient.GetUserGroups(userId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while retrieving access group %s", err), http.StatusInternalServerError)
		}

		groupToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"issuer":      userId,
			"permissions": res,
			"exp":         time.Now().Add(time.Second * config.JWTExpirationTime()).Unix(), // expires in 1 hour
		})

		groupTokenString, err := groupToken.SignedString(config.JWTSecret())
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

func checkHTTPReqAuth(r *http.Request, wasmPool *wasm.ExecutorPool) (string, *HttpError) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "ERROR_ID", &NoAuthHeaderErr
	}

	if !strings.HasPrefix(authHeader, AuthPrefix) {
		return "ERROR_ID", &InvalidAuthHeaderErr
	}

	authToken := strings.TrimPrefix(authHeader, AuthPrefix)
	executor := wasmPool.Acquire()
	userId, fnErrCode := executor.InvokeAuthFn("auth", authToken)
	executor.Release()

	if fnErrCode != 200 {
		return "ERROR_ID", &AuthFailedErr
	}

	return userId, nil
}
