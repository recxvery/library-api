package http

import (
	"encoding/json"
	"log"
	"net/http"
	"rest-api-app/books"

	"github.com/gorilla/mux"
)

type HTTPHandlers struct {
	lib *books.Lib
}

func NewHTTPHandlers(lib *books.Lib) *HTTPHandlers {
	return &HTTPHandlers{
		lib: &books.Lib{},
	}
}

func (h *HTTPHandlers) HandlerGetBookInfo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	book, err := h.lib.GetBookInfo(id)
	if err != nil {
		http.Error(w, string(errToJSON(err)), http.StatusNotFound)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		log.Println(err)
		http.Error(w, string(errToJSON(err)), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusFound)
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
		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
		log.Println(err)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(data)); err != nil {
		log.Println(err)
		return
	}
}


func (h *HTTPHandlers) HandlerGetBooks(w http.ResponseWriter, r *http.Request) {
	books := h.lib.GetAllBooks()

	data, err := json.MarshalIndent(books, "", "	")
	if err != nil {
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(data)); err != nil {
		log.Println(err)
		return
	}
}

func (h *HTTPHandlers) HandlerGetBooksByAuthor(w  http.ResponseWriter, r *http.Request) {
	author := mux.Vars(r)["author"]

	books := 
}