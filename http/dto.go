package http

import (
	"encoding/json"
	"rest-api-app/db"
	"time"
)

type BookDTO struct {
	Author string `json:"author"`
	Title  string `json:"title"`
	Description string `json:"description"`
	Pages  int    `json:"pages"`
	Year   int    `json:"year"`
}

type errDTO struct {
	Message string
	Time    time.Time
}

func errToJSON(err error) []byte {
	jsonErr, _ := json.MarshalIndent(errDTO{
		Message: err.Error(),
		Time:    time.Now(),
	}, "", "	")

	return jsonErr
}

func ValidateToCreate(book BookDTO) (db.BookModel, error) {
	if book.Title != "" && book.Author != "" && book.Pages > 0 && book.Year > 0 {
		return db.BookModel{
			Title:      book.Title,
			Author:     book.Author,
			Description: &book.Description,
			PagesCount: book.Pages,
			Year:       book.Year,
			AddedAt:    time.Now(),
		}, nil
	}

	return db.BookModel{}, ErrorNotEnoughData
}
