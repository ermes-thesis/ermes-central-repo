package main

import (
	"centralRepoDemo/db"
	"centralRepoDemo/db/DAO"
	"centralRepoDemo/gateway"
	"log"
	"net/http"
)

func main() {
	pool, err := db.NewPool()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	daoFactory, err := DAO.NewDAOFactory(pool)

	if err != nil {
		log.Fatal(err)
	}
	router := gateway.NewRouter(daoFactory)

	log.Printf("HTTP Gateway starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
