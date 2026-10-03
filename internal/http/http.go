package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"rest-api-app/internal/core"
	"rest-api-app/internal/db"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type HTTPHandlers struct {
	repo           db.BookRepository
	handlersLogger *zap.Logger
}

func NewHTTPHandlers(db db.BookRepository, logg *zap.Logger) *HTTPHandlers {
	return &HTTPHandlers{
		repo:           db,
		handlersLogger: logg,
	}
}

func (h *HTTPHandlers) HandlerGetBookInfo(w http.ResponseWriter, r *http.Request) {
	idFromQuery := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idFromQuery)

	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	book, err := h.repo.GetBook(r.Context(), id) 
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		h.handlersLogger.Error(fmt.Sprint("write to user err: ", err.Error()))
		return
	}
}

func (h *HTTPHandlers) HandlerAddNewBook(w http.ResponseWriter, r *http.Request) {
	var inputBook BookDTO

	err := json.NewDecoder(r.Body).Decode(&inputBook)
	if err != nil {
		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
		return
	}

	book, err := ValidateToCreate(inputBook)
	if err != nil {
		h.handlersLogger.Error(fmt.Sprint("validate to create handler err: ", err.Error()))
		return
	}
	savedBook, err := h.repo.InsertBook(r.Context(), book)
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	data, err := json.MarshalIndent(savedBook, "", "		")
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(data)); err != nil {
		h.handlersLogger.Error(fmt.Sprint("write to user err: ", err.Error()))
		return
	}
}

func (h *HTTPHandlers) HandlerGetBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query() //получаем все query параметры

	author := query.Get("author")

	var read *bool //if nil значит параметр не задан для read

	if readStr := query.Get("read"); readStr != "" {
		value, err := strconv.ParseBool(readStr)
		if err != nil {
			errorHandle(w, err, h.handlersLogger)
			return
		}

		read = &value //передаем значение
	}

	books, err := h.repo.GetBooks(r.Context(), author, read)
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	data, err := json.MarshalIndent(books, "", "	")
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(data)); err != nil {
		h.handlersLogger.Error(fmt.Sprint("write to user err: ", err.Error()))
		return
	}
}

func (h *HTTPHandlers) HandlerRemoveBook(w http.ResponseWriter, r *http.Request) {
	idFromQuery := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idFromQuery)

	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	book, err := h.repo.DeleteBook(r.Context(), id)
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(data)); err != nil {
		h.handlersLogger.Error(fmt.Sprint("write to user err: ", err.Error()))
		return
	}
}

func (h *HTTPHandlers) HandlerMakeBookRead(w http.ResponseWriter, r *http.Request) {
	idFromQuery := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idFromQuery)

	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	book, err := h.repo.UpdateBook(r.Context(), id)
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		errorHandle(w, err, h.handlersLogger)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(data); err != nil {
		h.handlersLogger.Error(fmt.Sprint("write to user err: ", err.Error()))
		return
	}
}

func errorHandle(w http.ResponseWriter, err error, log *zap.Logger) {
	log.Error(err.Error())

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		http.Error(w, string(errToJSON(err)), http.StatusNotFound)
	case errors.Is(err, strconv.ErrSyntax) || errors.Is(err, strconv.ErrRange):
		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
	case errors.Is(err, core.ErrorNotEnoughData):
		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
	default:
		http.Error(w, string(errToJSON(err)), http.StatusInternalServerError)
	}
}
