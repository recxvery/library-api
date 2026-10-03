package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type BookRepository interface {
	InsertBook(ctx context.Context, Book BookModel) (BookModel, error)
	UpdateBook(ctx context.Context, id int) (BookModel, error)
	DeleteBook(ctx context.Context, id int) (BookModel, error)
	GetBook(ctx context.Context, id int)  (BookModel, error)
	GetBooks(ctx context.Context, author string, read *bool) ([]BookModel, error)
}

type DbRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewBookRepo(pool_connection *pgxpool.Pool, Logg *zap.Logger) *DbRepository {
	return &DbRepository{
		pool:   pool_connection,
		logger: Logg,
	}
}

type BookModel struct {
	ID          int
	Title       string
	Author      string
	Description *string
	Year        int
	PagesCount  int
	IsRead      bool
	AddedAt     time.Time
	ReadAt      *time.Time
}

func ConnectDB(ctx context.Context) (*pgxpool.Pool, error) {
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

/* func (r *DbRepository) CreateDB(ctx context.Context) error {
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

	if _, err := r.pool.Exec(ctx, sqlQuery); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
migrations used
*/
