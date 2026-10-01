package http

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"time"
)

// структура для имитации writera
type ResponseWriter struct {
	responseWrite  http.ResponseWriter
	statusCode     int
	statusRecieved bool //флажок на то отправили ли мы HTTP-статус
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
	if rp.statusRecieved == false {
		rp.statusRecieved = true
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

func Middleware(next http.Handler) http.Handler {
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

		log.Printf("%s %s %v %s IP: %s User-agent: %s ", r.Method, r.URL.Path, newWriter.statusCode, requestTime, ip, user_agent)
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close") 
				http.Error(
					w,
					"Internal Server Error",
					500,
				)
				log.Println(err)
				debug.PrintStack() 
			}
		}()

		next.ServeHTTP(w, r)
	})
}
