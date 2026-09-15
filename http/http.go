package http

import (
	"encoding/json"
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

func (h *HTTPHandlers) HandlerGetBooksInfo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	book, err := h.lib.GetBookInfo(id)
	if err != nil {
		http.Error(w, string(errToJSON(err)), http.StatusNotFound)
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		http.Error(w, string(errToJSON(err)), http.StatusInternalServerError)
	}

	w.Write(data)
	w.WriteHeader(http.StatusFound)
}


func (h *HTTPHandlers) HandlerAddNewBook(w http.ResponseWriter, r *http.Request) {
	var inputBook BookDTO

	err := json.NewDecoder(r.Body).Decode(&inputBook)
	if err != nil {
		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
	}

	book := h.lib.AddBook(inputBook)
}
