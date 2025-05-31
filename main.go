package main

import (
	"centralRepoDemo/db"
	"centralRepoDemo/gateway"
	"fmt"
	"github.com/ermes-thesis/ermes-wasm-test/wasm-pool"
	"log"
	"net/http"
)

func main() {
	runWithRedis()
}

func runWithRedis() {
	rdClient := db.NewRedisClient()
	wasmPool := wasm.NewExecutorPool(10)

	defer func(rdClient *db.RedisClient) {
		err := rdClient.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rdClient)

	err := rdClient.AddGroupToUser("DOCTORS", "1")
	if err != nil {
		log.Fatal(err)
	}
	groups, err := rdClient.GetUserGroups("1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The groups for user:%d, are: %v\n", 1, groups)

	err = rdClient.CreateGroup("PROVA")
	if err != nil {
		log.Fatal(err)
	}
	existingGroups, err := rdClient.GetExistingGroups()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The existing groups are: %v\n", existingGroups)
	err = rdClient.DeleteGroup("PROVA")
	if err != nil {
		log.Fatal(err)
	}
	existingGroups, err = rdClient.GetExistingGroups()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The existing groups are: %v\n", existingGroups)

	router := gateway.NewRouter(rdClient, wasmPool)

	log.Printf("HTTP Gateway starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
