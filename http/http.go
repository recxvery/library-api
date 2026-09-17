package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"rest-api-app/books"
	"strconv"

	"github.com/gorilla/mux"
)

type HTTPHandlers struct {
	lib *books.Lib
}

func NewHTTPHandlers(lib *books.Lib) *HTTPHandlers {
	return &HTTPHandlers{
		lib: lib,
	}
}

func (h *HTTPHandlers) HandlerGetBookInfo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if len(id) == 0 {
		http.Error(w, string(errToJSON(errors.New("Len 0 is imposibble for id"))), http.StatusBadRequest)
		return
	}

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
		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
		log.Println(err)
		return
	}

	savedBook := h.lib.AddBook(book)

	data, err := json.MarshalIndent(savedBook, "", "		")
	if err != nil {
		http.Error(w, string(errToJSON(err)), http.StatusInternalServerError)
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
	query := r.URL.Query() //получаем все query параметры

	author := query.Get("author")

	var read *bool //if nil значит параметр не задан для read

	if readStr := query.Get("read"); readStr != "" {
		value, err := strconv.ParseBool(readStr)
		if err != nil {
			http.Error(w, string(errToJSON(errors.New("must be true or false"))), http.StatusBadRequest)
			log.Println(err)
			return
		}

		read = &value //передаем значение
	}

	books := h.lib.FilterBooks(author, read)

	data, err := json.MarshalIndent(books, "", "	")
	if err != nil {
		http.Error(w, string(errToJSON(err)), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(data)); err != nil {
		log.Println(err)
		return
	}
}

// func (h *HTTPHandlers) HandlerGetBooksByAuthor(w http.ResponseWriter, r *http.Request) {
// 	author := r.URL.Query().Get("author")

// 	if len(author) <= 1 {
// 		err := errors.New("Author being empty is not possible")
// 		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
// 		return
// 	}

// 	books := h.lib.GetBooksByAuthor(author)

// 	data, err := json.MarshalIndent(books, "", "	")
// 	if err != nil {
// 		http.Error(w, string(errToJSON(err)), http.StatusInternalServerError)
// 		log.Println(err)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	if _, err := w.Write([]byte(data)); err != nil {
// 		log.Println(err)
// 		return
// 	}
// }

// func (h *HTTPHandlers) HandlerGetReadBooks(w http.ResponseWriter, r *http.Request) {
// 	var (
// 		isRead       bool
// 		booksByParam map[string]books.Book
// 		err          error
// 	)
// 	isReadNotBool := strings.ToLower(r.URL.Query().Get("read"))

// 	isRead, err = strconv.ParseBool(isReadNotBool)
// 	if err != nil {
// 		http.Error(w, string(errToJSON(err)), http.StatusBadRequest)
// 		return
// 	}

// 	if isRead {
// 		booksByParam = h.lib.GetReadTrueBooks()
// 	} else {
// 		booksByParam = h.lib.GetUnreadBooks()
// 	}

// 	data, err := json.MarshalIndent(booksByParam, "", "		")
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	if _, err = w.Write([]byte(data)); err != nil {
// 		log.Println(err)
// 		return
// 	}
// }

func (h *HTTPHandlers) HandlerRemoveBook(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if len(id) == 0 {
		http.Error(w, string(errToJSON(errors.New("Len 0 is imposibble for id"))), http.StatusBadRequest)
		return
	}

	book, err := h.lib.RemoveBook(id)
	if err != nil {
		http.Error(w, string(errToJSON(err)), http.StatusNotFound)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
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

func (h *HTTPHandlers) HandlerMakeBookRead(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if len(id) == 0 {
		http.Error(w, string(errToJSON(errors.New("Len 0 is imposibble for id"))), http.StatusBadRequest)
		return
	}

	book, err := h.lib.ReadSome(id)
	if err != nil {
		http.Error(w, string(errToJSON(err)), http.StatusNotFound)
		log.Println(err)
		return
	}

	data, err := json.MarshalIndent(book, "", "		")
	if err != nil {
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(data); err != nil {
		log.Println(err)
		return
	}
}
