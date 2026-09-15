package http

import (
	"encoding/json"
	"rest-api-app/books"
	"time"
)

type BookDTO struct {
	Author string
	Title string
	Pages int
	Year int
}

type errDTO struct {
	Message string
	Time time.Time
}

func errToJSON(err error) []byte {
	jsonErr, _ := json.MarshalIndent(errDTO{
		Message: err.Error(),
		Time: time.Now(),
	}, "", "	")

	return jsonErr
}

func ValidateToCreate(book BookDTO) (books.Book, error) {
	if book.Title != "" && book.Author != "" && book.Pages <= 0 && book.Year <= 0 {
		return books.Book{
			Title:      book.Title,
			Author:     book.Author,
			PagesCount: book.Pages,
			Year:       book.Year,
			AddedAt:    time.Now(),
		}, nil
	}

	return books.Book{}, books.ErrorNotEnoughData
}

