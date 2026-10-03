package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"rest-api-app/internal/core"
	"rest-api-app/internal/db"
	"rest-api-app/internal/http"

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

	logger, closeLogger, err := core.NewLogger("INFO")
	if err != nil {
		log.Fatal(err)
	}

	defer closeLogger()

	repo := db.NewBookRepo(pool, logger)

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

	handlers := http.NewHTTPHandlers(repo, logger)
	server := http.NewServer(handlers, logger)

	// log.Println("Server started")
	if err := server.StartServer(); err != nil {
		server.ServerLogger.Error(fmt.Sprint("Server shut down with trouble", err.Error()))
	} else {
		log.Println("Gracefully shutodwn completed successfully")
	}
}
