package handlers

import (
	"centralRepoDemo/db"
	"github.com/ermes-thesis/ermes-wasm-test/wasm-pool"
	"net/http"
)

func RemoveFromGroupHandler(redisClient *db.RedisClient, wasmPool *wasm.ExecutorPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//userId, httpErr := CheckHTTPReqAuth(r, wasmPool)
		//if httpErr != nil {
		//	http.Error(w, httpErr.Error(), httpErr.StatusCode())
		//}
		userId := r.URL.Query().Get("userId")
		groupName := r.URL.Query().Get("group")
		err := redisClient.RemoveGroupFromUser(groupName, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
