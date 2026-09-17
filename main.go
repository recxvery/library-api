package main

import (
	"log"
	"rest-api-app/books"
	"rest-api-app/http"
)

func main() {
	newLib := books.NewLib()
	handlers := http.NewHTTPHandlers(newLib)
	server := http.NewServer(handlers)

	log.Println("Server started")
	if err := server.StartServer(); err != nil {
		log.Println(err.Error())
	}
}