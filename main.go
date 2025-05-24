package main

import (
	"centralRepoDemo/db/postgres"
	"centralRepoDemo/db/postgres/DAO"
	"centralRepoDemo/db/redis"
	"centralRepoDemo/gateway"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var uInput int
	for ok := true; ok; ok = uInput <= 0 || uInput > 3 {
		fmt.Println("DB to use:")
		fmt.Println("1) Postgres")
		fmt.Println("2) Redis")
		fmt.Println("3) Exit")

		_, err := fmt.Scanf("%d", &uInput)
		if err != nil {
			return
		}
	}

	switch uInput {
	case 1:
		runWithPostgres()
	case 2:
		runWithRedis()
	case 3:
		os.Exit(0)
	}

}

func runWithPostgres() {
	pool, err := postgres.NewPostgresPool()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	daoFactory, err := DAO.NewDAOFactory(pool)

	if err != nil {
		log.Fatal(err)
	}
	router := gateway.NewRouterWithPostgres(daoFactory)

	log.Printf("HTTP Gateway starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func runWithRedis() {
	rdClient := redis.NewRedisClient()
	defer func(rdClient *redis.RedisClient) {
		err := rdClient.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rdClient)

	err := rdClient.AddGroupToUser("DOCTORS", 1)
	if err != nil {
		log.Fatal(err)
	}
	groups, err := rdClient.GetUserGroups(1)
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

	router := gateway.NewRouterWithRedis(rdClient)

	log.Printf("HTTP Gateway starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
