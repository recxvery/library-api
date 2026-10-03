package http

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// структура для имитации writera
type ResponseWriter struct {
	responseWrite  http.ResponseWriter
	statusCode     int
	statusReceived bool //флажок на то отправили ли мы HTTP-статус
}

// constructor
func NewResponsewriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		responseWrite: w,
		statusCode:    200,
	}
}

// ниже копируем методы чтобы удовлетворять контракту интерфейса
func (rp *ResponseWriter) WriteHeader(statusCode int) {
	//если статус уже получен мы не имеем права ег оменять
	if rp.statusReceived == false {
		rp.statusReceived = true
		rp.statusCode = statusCode
		rp.responseWrite.WriteHeader(statusCode)
	}
}

func (rp *ResponseWriter) Write(s []byte) (int, error) {
	data, err := rp.responseWrite.Write(s)
	if err != nil {
		return 0, err
	}

	return data, nil
}

// хендлер будет работать с настоящим Header'
func (rp *ResponseWriter) Header() http.Header {
	return rp.responseWrite.Header()
}

func (app *AppServer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newWriter := NewResponsewriter(w) //cоздаём новый writer через котоырй будем протаскивать наш статус код
		measureTime := time.Now()         //время для лога

		remoteaddr := r.RemoteAddr //ip для лога

		ip, _, err := net.SplitHostPort(remoteaddr)
		if err != nil {
			ip = remoteaddr //если порт не сплитнулся то берём полный ip address
		}
		user_agent := r.UserAgent()

		next.ServeHTTP(newWriter, r) //вызываем следующий хендлер

		requestTime := fmt.Sprintf("%v ms", time.Since(measureTime).Milliseconds())

		app.ServerLogger.Info(fmt.Sprintf("%s %s %v %s IP: %s User-agent: %s ", r.Method, r.URL.Path, newWriter.statusCode, requestTime, ip, user_agent))
	})
}

func (app *AppServer) RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				http.Error(
					w,
					"Internal Server Error",
					500,
				)
				app.ServerLogger.Error(
					"panic revocered",
					zap.Any("panic:", err),
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *AppServer) ShutdownMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.shuttingDown.Load() {
			http.Error(w,
				"The server is undergoing a graceful shutdown. Active requests are being processed, but no new requests are accepted. Please try again later.",
				http.StatusServiceUnavailable,
			)
			app.ServerLogger.Warn("user try to make request while graceful shutdown")

			return
		}
		next.ServeHTTP(w, r)
	})
}
