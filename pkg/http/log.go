package http

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

type logger struct {
	http.Handler
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	l.Handler.ServeHTTP(lrw, r)
	log.Print(r.URL, lrw.statusCode)
}

type file []byte

func (f file) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write(f)
}
