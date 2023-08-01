package server

import (
	"handler"
	"net/http"
)

const (
	CONN_HOST = ""
	CONN_PORT = "8080"
)

func webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handler.ShortenURL(w, r)
	} else if r.Method == http.MethodGet {
		handler.GetFullUrl(w, r)
	} else {
		// allow only post/get requests, otherwise send a 405 code
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func Run() error {
	return http.ListenAndServe(CONN_HOST+":"+CONN_PORT, http.HandlerFunc(webhook))
}
