package handlers

import (
	"centralRepoDemo/db"
	"github.com/ermes-thesis/ermes-wasm-test/wasm-pool"
	"net/http"
)

func DeleteGroupHandler(redisClient *db.RedisClient, wasmPool *wasm.ExecutorPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, httpErr := CheckHTTPReqAuth(r, wasmPool)
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
