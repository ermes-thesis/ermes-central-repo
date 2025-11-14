package handlers

import (
	"encoding/json"
	"ermes-central-repo/caches"
	"ermes-central-repo/db"
	"ermes-central-repo/tokenEncoder"
	"fmt"
	"net/http"
	"strings"

	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/ermes-thesis/ermes-common/wasm-pool"
)

func RetrieveGroupToken(rdbClient db.GenericDB, wasmPool wasm.ExecutorPool, tokenEncoder tokenEncoder.TokenEncoder, activeAuthFn *string, log logger.Logger, cache caches.FunctionCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		userId, httpErr := checkHTTPReqAuth(r, wasmPool, activeAuthFn, cache)
		if httpErr != nil {
			log.WarnfWith("Authorization failed while executing '%s' authentication function",
				[]any{*activeAuthFn},
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"error", httpErr.Error(),
			)
			log.DebugfWith("Authorization failed while executing '%s' authentication function",
				[]any{*activeAuthFn},
				"fn_error_code", httpErr.StatusCode(),
				"fn_error_message", httpErr.Error(),
				"headers", r.Header,
			)
			http.Error(w, httpErr.Error(), httpErr.StatusCode())
			return
		}

		groups, err := rdbClient.GetUserGroups(userId)
		if err != nil {
			log.Warn("Failed to retrieve user groups",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"error", err,
			)
			log.Debug("Failed to retrieve user groups",
				"user_id", userId,
				"headers", r.Header,
			)
			http.Error(w, fmt.Sprintf("Could not retrieve groups for user '%s': %v", userId, err), http.StatusInternalServerError)
			return
		}

		groupTokenString, err := tokenEncoder.EncodeGroupToken(groups, userId)
		if err != nil {
			log.Warn("Failed to encode group token",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"error", err,
			)
			log.Debug("Failed to encode group token",
				"user_id", userId,
				"groups", groups,
				"headers", r.Header,
			)
			http.Error(w, fmt.Sprintf("Failed to generate group token: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"group_token": groupTokenString,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Warn("Failed to write group token response",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"error", err,
			)
			log.Debug("Failed to write group token response",
				"user_id", userId,
				"groups", groups,
				"group_token_string", groupTokenString,
				"headers", r.Header,
			)
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
			return
		}

		log.InfofWith("Group token generated successfully for '%s' user",
			[]any{userId},
			"remote_addr", r.RemoteAddr,
			"method", r.Method,
			"path", r.URL.RequestURI(),
			"user_id", userId,
			"groups", groups,
		)
	}
}

func checkHTTPReqAuth(r *http.Request, wasmPool wasm.ExecutorPool, activeAuthFn *string, cache caches.FunctionCache) (string, *HttpError) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "ERROR_ID", &NoAuthHeaderErr
	}

	if !strings.HasPrefix(authHeader, AuthPrefix) {
		return "ERROR_ID", &InvalidAuthHeaderErr
	}

	authToken := strings.TrimPrefix(authHeader, AuthPrefix)

	if *activeAuthFn == "" {
		return "ERROR_ID", &InvalidAuthFunctionErr
	}

	authFn := strings.SplitN(*activeAuthFn, "@", 2)

	wasmBinary, err := cache.GetBinary(authFn[0], authFn[1])
	if err != nil {
		return "ERROR_ID", &InvalidAuthFunctionErr
	}

	response := wasmPool.Invoke(wasmBinary, "authEntry", []byte(authToken))

	if response.HttpCode != 200 {
		return "ERROR_ID", &AuthFailedErr
	}

	return response.Response.(string), nil
}
