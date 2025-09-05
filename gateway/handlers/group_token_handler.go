package handlers

import (
	"encoding/json"
	"ermes-central-repo/db"
	"ermes-central-repo/tokenEncoder"
	"fmt"
	"github.com/ermes-thesis/ermes-wasm-pool/wasm-pool"
	"net/http"
	"strings"
)

func RetrieveGroupToken(rdbClient db.GenericDB, wasmPool *wasm.ExecutorPool, tokenEncoder tokenEncoder.TokenEncoder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.Body.Close()
		userId, httpErr := checkHTTPReqAuth(r, wasmPool)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode())
			return
		}
		groups, err := rdbClient.GetUserGroups(userId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while retrieving access group %s", err), http.StatusInternalServerError)
			return
		}

		groupTokenString, err := tokenEncoder.EncodeGroupToken(groups, userId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while encoding access group token %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"group_token": groupTokenString,
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
