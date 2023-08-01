package main

import (
	"fmt"
	"handler"
	"net/http"
	"storage"
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

func run() error {
	return http.ListenAndServe(CONN_HOST+":"+CONN_PORT, http.HandlerFunc(webhook))
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}

	err := storage.Load()
	if err != nil {
		fmt.Println(err)
	}

}
