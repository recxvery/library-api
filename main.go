package main

import (
	"context"
	"log"
	"rest-api-app/db"
	"rest-api-app/http"
)

func main() {
	ctx := context.Background()

	conn, err := db.ConnectDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx) 

	repo := db.NewBookRepo(conn)

	if err := repo.CreateDB(ctx); err != nil {
		log.Fatal(err)
	}

	handlers := http.NewHTTPHandlers(repo)
	server := http.NewServer(handlers)

	log.Println("Server started")
	if err := server.StartServer(); err != nil {
		log.Println(err.Error())
	}
}


