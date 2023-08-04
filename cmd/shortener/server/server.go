package server

import (
	"handler"
	"net/http"
	"storage"
)

const (
	CONN_HOST = ""
	CONN_PORT = "8080"
)

var hndlr *handler.Handler

func webhook(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		hndlr.ShortenURL(w, r)
	} else if r.Method == http.MethodGet {
		hndlr.GetFullUrl(w, r)
	} else {
		// allow only post/get requests, otherwise send a 405 code
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func Run() error {
	memoryStorage := storage.NewMemoryStorage()
	hndlr = handler.NewHandler(memoryStorage)
	return http.ListenAndServe(CONN_HOST+":"+CONN_PORT, http.HandlerFunc(webhook))
}
