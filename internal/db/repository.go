package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/k0kubun/pp"
)

func (r *DbRepository) InsertBook(ctx context.Context, Book BookModel) (BookModel, error) {
	sqlQuery := `
	INSERT INTO books (book_title, book_author, publish_year, pages_count, book_read, added_at,  book_description)
	VALUES ($1, $2, $3, $4, $5, $6, $7) 
	RETURNING id, book_title, book_author, publish_year, pages_count, book_read, added_at, read_at, book_description;
	`

	var result BookModel

	err := r.pool.QueryRow(ctx,
		sqlQuery,
		Book.Title,
		Book.Author,
		Book.Year,
		Book.PagesCount,
		Book.IsRead,
		Book.AddedAt,
		Book.Description).Scan(
		&result.ID,
		&result.Title,
		&result.Author,
		&result.Year,
		&result.PagesCount,
		&result.IsRead,
		&result.AddedAt,
		&result.ReadAt,
		&result.Description,
	)

	if err != nil {
		r.logger.Error(fmt.Sprint("insert row some", err.Error()))
		return BookModel{}, err
	}

	return result, nil
}

func (r *DbRepository) UpdateBook(ctx context.Context, id int) (BookModel, error) {
	sqlQuery := `
	UPDATE books 
	SET book_read=$1, read_at=$2
	WHERE id=$3
	RETURNING id, book_title, book_author, publish_year, pages_count, book_read, added_at, read_at, book_description;
	`
	var result BookModel

	err := r.pool.QueryRow(ctx, sqlQuery, true, time.Now(), id).Scan(
		&result.ID,
		&result.Title,
		&result.Author,
		&result.Year,
		&result.PagesCount,
		&result.IsRead,
		&result.AddedAt,
		&result.ReadAt,
		&result.Description,
	)

	if err != nil {
		r.logger.Error(fmt.Sprint("update book query err", err.Error()))
		return BookModel{}, err
	}

	return result, nil
}

func (r *DbRepository) DeleteBook(ctx context.Context, id int) (BookModel, error) {
	sqlQuery := `
		DELETE FROM books WHERE id=$1
		RETURNING id, book_title, book_author, publish_year, pages_count, book_read, added_at, read_at, book_description;
	`
	var result BookModel

	err := r.pool.QueryRow(ctx, sqlQuery, id).Scan(
		&result.ID,
		&result.Title,
		&result.Author,
		&result.Year,
		&result.PagesCount,
		&result.IsRead,
		&result.AddedAt,
		&result.ReadAt,
		&result.Description,
	)

	if err != nil {
		r.logger.Error(fmt.Sprint("delete book query err", err.Error()))
		return BookModel{}, err
	}

	return result, nil
}

func (r *DbRepository) GetBook(ctx context.Context, id int) (BookModel, error) {
	sqlQuery := `
		SELECT id, book_title, book_author, publish_year, pages_count, book_read,added_at, read_at, book_description FROM books
		WHERE id=$1;
	`

	var result BookModel

	err := r.pool.QueryRow(ctx, sqlQuery, id).Scan(
		&result.ID,
		&result.Title,
		&result.Author,
		&result.Year,
		&result.PagesCount,
		&result.IsRead,
		&result.AddedAt,
		&result.ReadAt,
		&result.Description,
	)

	if err != nil {
		r.logger.Error(err.Error())
		return BookModel{}, err
	}

	return result, nil
}

func (r *DbRepository) GetBooks(ctx context.Context, author string, read *bool) ([]BookModel, error) {
	sqlQuery := `
	SELECT id, book_title, book_author, publish_year, pages_count, book_read,added_at, read_at, book_description FROM books	
	`

	var conditions []string //для динамического изменения SQL-запроса
	var args []any          //для распаковки зачений в conn.Query()

	if author != "" { //проверяем наличие автора
		conditions = append(conditions, fmt.Sprintf("book_author=$%d", len(args)+1))
		args = append(args, author)
	}

	if read != nil {
		conditions = append(conditions, fmt.Sprintf("book_read=$%d", len(args)+1))
		args = append(args, *read)
	}

	if len(conditions) > 0 {
		sqlQuery += ` WHERE ` + strings.Join(conditions, " AND ")
	}
	sqlQuery += ` ORDER BY id ASC;`

	rows, err := r.pool.Query(ctx, sqlQuery, args...)
	if err != nil {
		r.logger.Error(fmt.Sprint("get books query err", err.Error()))
		return []BookModel{}, err
	}

	defer rows.Close()

	var books []BookModel

	for rows.Next() {
		var Book BookModel

		err := rows.Scan(
			&Book.ID,
			&Book.Title,
			&Book.Author,
			&Book.Year,
			&Book.PagesCount,
			&Book.IsRead,
			&Book.AddedAt,
			&Book.ReadAt,
			&Book.Description,
		)

		if err != nil {
			return []BookModel{}, err
		}

		books = append(books, Book)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(fmt.Sprint("Get books row err", err.Error()))
		return []BookModel{}, rows.Err()
	}

	r.logger.Info(pp.Sprint("books fetched", books)) //for logs
	return books, nil
}

// func (r *DbRepository) GetBooks(ctx context.Context) ([]BookModel, error) {
// 	sqlQuery := `
// 	SELECT id, book_title, book_author, publish_year, pages_count, book_read, added_at, read_at
// 	FROM BOOKS
// 	ORDER BY id ASC;
// 	`
// 	rows, err := r.conn.Query(ctx, sqlQuery)
// 	if err != nil {
// 		return []BookModel{}, err
// 	}

// 	defer rows.Close()

// 	books := []BookModel{}

// 	for rows.Next() {
// 		var Book BookModel

// 		err := rows.Scan(
// 			&Book.ID,
// 			&Book.Title,
// 			&Book.Author,
// 			&Book.Year,
// 			&Book.PagesCount,
// 			&Book.IsRead,
// 			&Book.AddedAt,
// 			&Book.ReadAt,
// 		)

// 		if err != nil {
// 			return []BookModel{}, err
// 		}

// 		books = append(books, Book)
// 	}

// 	pp.Print(books)
// 	return books, nil
// }
