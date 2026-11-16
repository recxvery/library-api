package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestResponseWriter(t *testing.T) {
	t.Run("defaults to status 200", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writer := NewResponsewriter(rec)

		if writer.statusCode != http.StatusOK {
			t.Errorf("status: got %d, want %d", writer.statusCode, http.StatusOK)
		}
	})

	t.Run("remembers the first status code", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writer := NewResponsewriter(rec)

		writer.WriteHeader(http.StatusCreated)
		writer.WriteHeader(http.StatusInternalServerError)

		if writer.statusCode != http.StatusCreated {
			t.Errorf("status: got %d, want %d", writer.statusCode, http.StatusCreated)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("recorder status: got %d, want %d", rec.Code, http.StatusCreated)
		}
	})

	t.Run("passes body and headers through", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writer := NewResponsewriter(rec)

		writer.Header().Set("X-Test", "yes")
		writer.WriteHeader(http.StatusAccepted)

		n, err := writer.Write([]byte("hello"))
		if err != nil {
			t.Fatalf("write: %v", err)
		}
		if n != len("hello") {
			t.Errorf("bytes written: got %d, want %d", n, len("hello"))
		}
		if rec.Body.String() != "hello" {
			t.Errorf("body: got %q, want %q", rec.Body.String(), "hello")
		}
		if rec.Header().Get("X-Test") != "yes" {
			t.Errorf("header: got %q, want %q", rec.Header().Get("X-Test"), "yes")
		}
	})
}

func TestRecoverMiddleware(t *testing.T) {
	app := NewServer(nil, zap.NewNop())

	handler := app.RecoverMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if body := rec.Body.String(); body != "Internal Server Error\n" {
		t.Errorf("body: got %q, want %q", body, "Internal Server Error\n")
	}
}

func TestRecoverMiddlewarePassesNormalResponses(t *testing.T) {
	app := NewServer(nil, zap.NewNop())

	handler := app.RecoverMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusTeapot)
	}
}

func TestShutdownMiddleware(t *testing.T) {
	t.Run("allows requests while running", func(t *testing.T) {
		app := NewServer(nil, zap.NewNop())

		called := false
		handler := app.ShutdownMiddleWare(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if !called {
			t.Error("next handler was not called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("rejects new requests during shutdown", func(t *testing.T) {
		app := NewServer(nil, zap.NewNop())
		app.shuttingDown.Store(true)

		called := false
		handler := app.ShutdownMiddleWare(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if called {
			t.Error("next handler must not be called during shutdown")
		}
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
	})
}

func TestLoggingMiddlewarePassesResponse(t *testing.T) {
	app := NewServer(nil, zap.NewNop())

	handler := app.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/books", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	req.Header.Set("User-Agent", "test-agent")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != "created" {
		t.Errorf("body: got %q, want %q", rec.Body.String(), "created")
	}
}
