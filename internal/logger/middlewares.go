package logger

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func LoggerMiddleware(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// эндпоинт /ping
			uri := r.RequestURI
			// метод запроса
			method := r.Method

			lw := loggingResponseWriter{
				ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			}
			next.ServeHTTP(&lw, r)

			duration := time.Since(start)

			log.Info().
				Str("Type", "Request").
				Str("Method", method).
				Str("URI", uri).
				Dur("Duration", duration).
				Msg("")

			log.Info().
				Str("Type", "Response").
				Int("StatusCode", lw.status).
				Int("Size", lw.size).
				Msg("")
		})
	}
}

type (
	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter

		status int
		size   int
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode // захватываем код статуса
}
