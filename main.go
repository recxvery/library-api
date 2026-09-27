package main

import (
	"context"
	"log"
	"os"
	"rest-api-app/db"
	"rest-api-app/http"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	ctx := context.Background()

	pool, err := db.ConnectDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := db.NewBookRepo(pool)

	m, err := migrate.New(
		"file://migrations",
		os.Getenv("CONN_STRING")+"?sslmode=disable",
	)

	if err != nil {
		log.Fatal(err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	handlers := http.NewHTTPHandlers(repo)
	server := http.NewServer(handlers)

	log.Println("Server started")
	if err := server.StartServer(); err != nil {
		log.Println(err.Error())
	}
}
