package httputil

import (
	"log"
	"net/http"
)

type loggingResponseWriter struct {
	http.ResponseWriter

	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

type Logger struct {
	http.Handler
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	l.Handler.ServeHTTP(lrw, r)
	log.Print(r.URL, lrw.statusCode)
}

type File []byte

func (f File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write(f)
}
