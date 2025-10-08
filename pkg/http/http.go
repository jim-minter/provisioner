package http

import (
	"net/http"
)

type HTTPServer struct {
	mux *http.ServeMux
}

func New() *HTTPServer {
	hs := &HTTPServer{
		mux: &http.ServeMux{},
	}

	hs.mux.Handle("/meta-data", file(nil))
	hs.mux.Handle("/vendor-data", file(nil))
	hs.mux.Handle("/user-data", file([]byte(``)))

	return hs
}

func (hs *HTTPServer) Serve() error {
	return http.ListenAndServe(":80", &logger{hs.mux})
}
