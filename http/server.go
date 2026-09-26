package http

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	httpHandlers *HTTPHandlers
}

func NewServer(handlers *HTTPHandlers) *HTTPServer {
	return &HTTPServer{
		httpHandlers: handlers,
	}
}

func (h *HTTPServer) StartServer() error {
	router := mux.NewRouter()

	router.Path("/books").Methods("POST").HandlerFunc(h.httpHandlers.HandlerAddNewBook)
	router.Path("/books").Methods("GET").HandlerFunc(h.httpHandlers.HandlerGetBooks)
	router.Path("/books/{id}").Methods("GET").HandlerFunc(h.httpHandlers.HandlerGetBookInfo)
	router.Path("/books/{id}").Methods("PATCH").HandlerFunc(h.httpHandlers.HandlerMakeBookRead)
	router.Path("/books/{id}").Methods("DELETE").HandlerFunc(h.httpHandlers.HandlerRemoveBook)

	if err := http.ListenAndServe(":8081", router); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}

	return nil
}
