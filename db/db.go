package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BookRepository struct {
	Pool *pgxpool.Pool
}

func NewBookRepo(pool_connection *pgxpool.Pool) *BookRepository {
	return &BookRepository{
		Pool: pool_connection,
	}
}

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

func ConnectDB(ctx context.Context) (*pgxpool.Pool, error) {
	// log.Println(conn_string)
	return pgxpool.New(ctx, os.Getenv("CONN_STRING"))
}

func CheckConnection(ctx context.Context, pool *pgxpool.Pool) {
	if err := pool.Ping(ctx); err != nil {
		log.Println(err)
		panic(err)
	} else {
		log.Println("succesfully connected")
	} //for /readyz
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

	if _, err := r.Pool.Exec(ctx, sqlQuery); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
