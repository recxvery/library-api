package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"rest-api-app/db"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type HTTPHandlers struct {
	repo *db.BookRepository
}

func NewHTTPHandlers(db *db.BookRepository) *HTTPHandlers {
	return &HTTPHandlers{
		repo: db,
	}
}

func (h *HTTPHandlers) HandlerGetBookInfo(w http.ResponseWriter, r *http.Request) {
	idFromQuery := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idFromQuery)

	if err != nil {
		errorHandle(w, err)
		return
	}

	book, err := h.repo.GetBook(r.Context(), id)
	if err != nil {
		errorHandle(w, err)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		errorHandle(w, err)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Println(err)
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
		errorHandle(w, err)
		return
	}
	savedBook, err := h.repo.InsertBook(r.Context(), book)
	if err != nil {
		errorHandle(w, err)
		return
	}

	data, err := json.MarshalIndent(savedBook, "", "		")
	if err != nil {
		errorHandle(w, err)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(data)); err != nil {
		log.Println(err)
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
			errorHandle(w, err)
			return
		}

		read = &value //передаем значение
	}

	books, err := h.repo.GetBooks(r.Context(), author, read)
	if err != nil {
		errorHandle(w, err)
		return
	}

	data, err := json.MarshalIndent(books, "", "	")
	if err != nil {
		errorHandle(w, err)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(data)); err != nil {
		log.Println(err)
		return
	}
}

func (h *HTTPHandlers) HandlerRemoveBook(w http.ResponseWriter, r *http.Request) {
	idFromQuery := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idFromQuery)

	if err != nil {
		errorHandle(w, err)
		return
	}

	book, err := h.repo.DeleteBook(r.Context(), id)
	if err != nil {
		errorHandle(w, err)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		errorHandle(w, err)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(data)); err != nil {
		log.Println(err)
		return
	}
}

func (h *HTTPHandlers) HandlerMakeBookRead(w http.ResponseWriter, r *http.Request) {
	idFromQuery := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idFromQuery)

	if err != nil {
		errorHandle(w, err)
		return
	}

	book, err := h.repo.UpdateBook(r.Context(), id)
	if err != nil {
		errorHandle(w, err)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		errorHandle(w, err)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(data); err != nil {
		log.Println(err)
		return
	}
}

func errorHandle(w http.ResponseWriter, err error) {
	log.Println(err)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		http.Error(w, string(errToJSON(err)), http.StatusNotFound)
	case errors.Is(err, strconv.ErrSyntax) || errors.Is(err, strconv.ErrRange):
		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
	case errors.Is(err, ErrorNotEnoughData):
		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
	default:
		http.Error(w, string(errToJSON(err)), http.StatusInternalServerError)
	}
}
