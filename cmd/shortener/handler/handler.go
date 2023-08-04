package handler

import (
	"io"
	"net/http"
	"shorturl"
	"storage"
	"strings"
)

type Handler struct {
	storage storage.Storage
}

func NewHandler(storage storage.Storage) *Handler {
	return &Handler{storage: storage}
}

// POST
func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	// установим правильный заголовок для типа данных
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")

	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}

	// get short url
	key, shurl := shorturl.Shorten(string(body), proto, r.Host)

	// save full url to storage with the key received earlier
	h.storage.Insert(key, string(body))

	// respond to client
	w.Header().Set("content-type", "text/plain")
	// set code 201
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shurl))
}

// GET
func (h *Handler) GetFullUrl(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	key = strings.Replace(key, "/", "", -1)
	if len(key) == 0 {
		// passed empty key
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// get full url from storage
	url, err := h.storage.Get(key)
	if err != nil || len(url) == 0 {
		// value not found for the passed key
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect) // 307
}
