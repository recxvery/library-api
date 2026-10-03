package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type AppServer struct {
	httpHandlers *HTTPHandlers
	shuttingDown atomic.Bool
}

func NewServer(handlers *HTTPHandlers) *AppServer {
	return &AppServer{
		httpHandlers: handlers,
	}
}

func (a *AppServer) StartServer() error {
	router := mux.NewRouter()
	router.Use(RecoverMiddleware)
	router.Use(a.ShutdownMiddleWare)
	router.Use(Middleware)

	router.Path("/books").Methods("POST").HandlerFunc(a.httpHandlers.HandlerAddNewBook)
	router.Path("/books").Methods("GET").HandlerFunc(a.httpHandlers.HandlerGetBooks)
	router.Path("/books/{id}").Methods("GET").HandlerFunc(a.httpHandlers.HandlerGetBookInfo)
	router.Path("/books/{id}").Methods("PATCH").HandlerFunc(a.httpHandlers.HandlerMakeBookRead)
	router.Path("/books/{id}").Methods("DELETE").HandlerFunc(a.httpHandlers.HandlerRemoveBook)

	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	return serverLifeCycle(a, server)
}

func serverLifeCycle(a *AppServer, server *http.Server) error {
	sigCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		log.Println("Server started at:", server.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-sigCtx.Done():
		a.shuttingDown.Store(true)
		ctxGS, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctxGS); err != nil {

			return err
		}
		return nil

	case err := <-errCh:
		return err
	}
}
