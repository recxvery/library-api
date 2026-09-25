package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

type BookModel struct {
	ID         int
	Title      string
	Author     string
	Year       int
	PagesCount int
	IsRead     bool
	AddedAt    time.Time
	ReadAt     *time.Time
}

func ConnectDB(ctx context.Context) (*pgx.Conn, error) {
	conn_string := os.Getenv("CONN_STRING")
	// log.Println(conn_string)

	return pgx.Connect(ctx, conn_string)
}

func CheckConnection(ctx context.Context, conn *pgx.Conn) {
	if err := conn.Ping(ctx); err != nil {
		log.Println(err)
		panic(err)
	} else {
		log.Println("succesfully connected")
	}
}

func (r *BookRepository) CreateDB(ctx context.Context) error {
	sqlQuery := `
	CREATE TABLE IF NOT EXISTS books (
		id SERIAL PRIMARY KEY,
		book_title VARCHAR(1000) NOT NULL,
		book_author VARCHAR(200) NOT NULL,
		publish_year INTEGER NOT NULL,
		pages_count 	INTEGER NOT NULL,
		book_read		BOOLEAN NOT NULL,
		added_at		TIMESTAMP NOT NULL,
		read_at			TIMESTAMP
	)`

	if _, err := r.conn.Exec(ctx, sqlQuery); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
