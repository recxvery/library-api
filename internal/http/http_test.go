package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"rest-api-app/internal/db"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type fakeRepo struct {
	InsertBookFunc func(ctx context.Context, book db.BookModel) (db.BookModel, error)
	UpdateBookFunc func(ctx context.Context, id int) (db.BookModel, error)
	DeleteBookFunc func(ctx context.Context, id int) (db.BookModel, error)
	GetBookFunc    func(ctx context.Context, id int) (db.BookModel, error)
	GetBooksFunc   func(ctx context.Context, author string, read *bool) ([]db.BookModel, error)
}

func (f *fakeRepo) InsertBook(ctx context.Context, book db.BookModel) (db.BookModel, error) {
	return f.InsertBookFunc(ctx, book)
}

func (f *fakeRepo) UpdateBook(ctx context.Context, id int) (db.BookModel, error) {
	return f.UpdateBookFunc(ctx, id)
}

func (f *fakeRepo) DeleteBook(ctx context.Context, id int) (db.BookModel, error) {
	return f.DeleteBookFunc(ctx, id)
}

func (f *fakeRepo) GetBook(ctx context.Context, id int) (db.BookModel, error) {
	return f.GetBookFunc(ctx, id)
}

func (f *fakeRepo) GetBooks(ctx context.Context, author string, read *bool) ([]db.BookModel, error) {
	return f.GetBooksFunc(ctx, author, read)
}

func newTestRouter(handlers *HTTPHandlers) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/books", handlers.HandlerAddNewBook).Methods(http.MethodPost)
	router.HandleFunc("/books", handlers.HandlerGetBooks).Methods(http.MethodGet)
	router.HandleFunc("/books/{id}", handlers.HandlerGetBookInfo).Methods(http.MethodGet)
	router.HandleFunc("/books/{id}", handlers.HandlerMakeBookRead).Methods(http.MethodPatch)
	router.HandleFunc("/books/{id}", handlers.HandlerRemoveBook).Methods(http.MethodDelete)
	return router
}

func newTestHandlers(repo db.BookRepository) (*HTTPHandlers, *mux.Router) {
	handlers := NewHTTPHandlers(repo, zap.NewNop())
	return handlers, newTestRouter(handlers)
}

func sampleBook(id int) db.BookModel {
	description := "A story about a journey"
	addedAt := time.Date(2026, time.May, 12, 10, 30, 0, 0, time.UTC)

	return db.BookModel{
		ID:          id,
		Title:       "The Hobbit",
		Author:      "J.R.R. Tolkien",
		Description: &description,
		Year:        1937,
		PagesCount:  310,
		IsRead:      false,
		AddedAt:     addedAt,
	}
}

func decodeBody[T any](t *testing.T, body []byte) T {
	t.Helper()

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("decode response body: %v (body: %s)", err, body)
	}
	return result
}

func assertBookEqual(t *testing.T, got, want db.BookModel) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("book mismatch:\ngot:  %+v\nwant: %+v", got, want)
	}
}

func assertNotFoundBody(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if !strings.Contains(rec.Body.String(), pgx.ErrNoRows.Error()) {
		t.Errorf("expected error body to contain %q, got %q", pgx.ErrNoRows.Error(), rec.Body.String())
	}
}

func TestHandlerGetBookInfo(t *testing.T) {
	t.Run("returns book by id", func(t *testing.T) {
		want := sampleBook(1)

		var receivedID int
		mockRepo := &fakeRepo{
			GetBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				receivedID = id
				return want, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/books/1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type: got %q, want %q", got, "application/json")
		}
		if receivedID != 1 {
			t.Errorf("repository received id %d, want 1", receivedID)
		}
		assertBookEqual(t, decodeBody[db.BookModel](t, rec.Body.Bytes()), want)
	})

	t.Run("returns bad request for non numeric id", func(t *testing.T) {
		mockRepo := &fakeRepo{
			GetBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				t.Fatal("repository must not be called")
				return db.BookModel{}, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/books/abc", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns not found when book does not exist", func(t *testing.T) {
		mockRepo := &fakeRepo{
			GetBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				return db.BookModel{}, pgx.ErrNoRows
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/books/42", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
		}
		assertNotFoundBody(t, rec)
	})

	t.Run("returns internal error when repository fails", func(t *testing.T) {
		mockRepo := &fakeRepo{
			GetBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				return db.BookModel{}, errors.New("connection lost")
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/books/1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestHandlerGetBooks(t *testing.T) {
	t.Run("returns all books without filters", func(t *testing.T) {
		want := []db.BookModel{sampleBook(1), sampleBook(2)}

		var receivedAuthor string
		var receivedRead *bool
		mockRepo := &fakeRepo{
			GetBooksFunc: func(ctx context.Context, author string, read *bool) ([]db.BookModel, error) {
				receivedAuthor = author
				receivedRead = read
				return want, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
		}
		if receivedAuthor != "" {
			t.Errorf("author filter: got %q, want empty", receivedAuthor)
		}
		if receivedRead != nil {
			t.Errorf("read filter: got %v, want nil", receivedRead)
		}

		got := decodeBody[[]db.BookModel](t, rec.Body.Bytes())
		if !reflect.DeepEqual(got, want) {
			t.Errorf("books mismatch:\ngot:  %+v\nwant: %+v", got, want)
		}
	})

	t.Run("passes author and read filters to repository", func(t *testing.T) {
		mockRepo := &fakeRepo{
			GetBooksFunc: func(ctx context.Context, author string, read *bool) ([]db.BookModel, error) {
				if author != "George Orwell" {
					t.Errorf("author: got %q, want %q", author, "George Orwell")
				}
				if read == nil || !*read {
					t.Errorf("read: got %v, want true", read)
				}
				return []db.BookModel{sampleBook(1)}, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/books?author=George+Orwell&read=true", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("returns bad request for invalid read filter", func(t *testing.T) {
		mockRepo := &fakeRepo{
			GetBooksFunc: func(ctx context.Context, author string, read *bool) ([]db.BookModel, error) {
				t.Fatal("repository must not be called")
				return nil, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/books?read=maybe", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns internal error when repository fails", func(t *testing.T) {
		mockRepo := &fakeRepo{
			GetBooksFunc: func(ctx context.Context, author string, read *bool) ([]db.BookModel, error) {
				return nil, errors.New("query failed")
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestHandlerAddNewBook(t *testing.T) {
	t.Run("creates book and returns it", func(t *testing.T) {
		saved := sampleBook(7)

		var inserted db.BookModel
		mockRepo := &fakeRepo{
			InsertBookFunc: func(ctx context.Context, book db.BookModel) (db.BookModel, error) {
				inserted = book
				return saved, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		body := []byte(`{
			"author": "J.R.R. Tolkien",
			"title": "The Hobbit",
			"description": "A story about a journey",
			"pages": 310,
			"year": 1937
		}`)

		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status: got %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		if inserted.Title != "The Hobbit" || inserted.Author != "J.R.R. Tolkien" {
			t.Errorf("unexpected inserted book: %+v", inserted)
		}
		if inserted.PagesCount != 310 || inserted.Year != 1937 {
			t.Errorf("unexpected inserted book: %+v", inserted)
		}
		if inserted.Description == nil || *inserted.Description != "A story about a journey" {
			t.Errorf("unexpected description: %v", inserted.Description)
		}
		assertBookEqual(t, decodeBody[db.BookModel](t, rec.Body.Bytes()), saved)
	})

	t.Run("returns bad request for malformed json", func(t *testing.T) {
		mockRepo := &fakeRepo{
			InsertBookFunc: func(ctx context.Context, book db.BookModel) (db.BookModel, error) {
				t.Fatal("repository must not be called")
				return db.BookModel{}, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(`{"title": `))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("does not save book with invalid data", func(t *testing.T) {
		mockRepo := &fakeRepo{
			InsertBookFunc: func(ctx context.Context, book db.BookModel) (db.BookModel, error) {
				t.Fatal("repository must not be called")
				return db.BookModel{}, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		body := []byte(`{"author": "", "title": "No author", "pages": 100, "year": 2000}`)
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code == http.StatusCreated {
			t.Error("book must not be created with invalid data")
		}
	})

	t.Run("returns internal error when repository fails", func(t *testing.T) {
		mockRepo := &fakeRepo{
			InsertBookFunc: func(ctx context.Context, book db.BookModel) (db.BookModel, error) {
				return db.BookModel{}, errors.New("insert failed")
			},
		}

		_, router := newTestHandlers(mockRepo)

		body := []byte(`{"author": "George Orwell", "title": "1984", "pages": 328, "year": 1949}`)
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestHandlerMakeBookRead(t *testing.T) {
	t.Run("marks book as read", func(t *testing.T) {
		want := sampleBook(3)
		want.IsRead = true

		var receivedID int
		mockRepo := &fakeRepo{
			UpdateBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				receivedID = id
				return want, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodPatch, "/books/3", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
		}
		if receivedID != 3 {
			t.Errorf("repository received id %d, want 3", receivedID)
		}
		assertBookEqual(t, decodeBody[db.BookModel](t, rec.Body.Bytes()), want)
	})

	t.Run("returns bad request for non numeric id", func(t *testing.T) {
		mockRepo := &fakeRepo{
			UpdateBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				t.Fatal("repository must not be called")
				return db.BookModel{}, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodPatch, "/books/abc", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns not found when book does not exist", func(t *testing.T) {
		mockRepo := &fakeRepo{
			UpdateBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				return db.BookModel{}, pgx.ErrNoRows
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodPatch, "/books/999", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
		}
		assertNotFoundBody(t, rec)
	})
}

func TestHandlerRemoveBook(t *testing.T) {
	t.Run("deletes book and returns it", func(t *testing.T) {
		want := sampleBook(5)

		var receivedID int
		mockRepo := &fakeRepo{
			DeleteBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				receivedID = id
				return want, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodDelete, "/books/5", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
		}
		if receivedID != 5 {
			t.Errorf("repository received id %d, want 5", receivedID)
		}
		assertBookEqual(t, decodeBody[db.BookModel](t, rec.Body.Bytes()), want)
	})

	t.Run("returns bad request for non numeric id", func(t *testing.T) {
		mockRepo := &fakeRepo{
			DeleteBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				t.Fatal("repository must not be called")
				return db.BookModel{}, nil
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodDelete, "/books/abc", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns not found when book does not exist", func(t *testing.T) {
		mockRepo := &fakeRepo{
			DeleteBookFunc: func(ctx context.Context, id int) (db.BookModel, error) {
				return db.BookModel{}, pgx.ErrNoRows
			},
		}

		_, router := newTestHandlers(mockRepo)

		req := httptest.NewRequest(http.MethodDelete, "/books/999", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestRouterReturnsNotFoundForUnknownRoute(t *testing.T) {
	mockRepo := &fakeRepo{}
	_, router := newTestHandlers(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
}
