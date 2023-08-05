package controllers

import (
	"io"
	"net/http"
	"strings"

	"github.com/wurt83ow/tinyurl/cmd/shortener/shorturl"
)

type Storage interface {
	Insert(k string, v string) error
	Get(k string) (string, error)
}

type BaseController struct {
	storage Storage
}

func NewBaseController(storage Storage) *BaseController {
	return &BaseController{storage: storage}
}

func (h *BaseController) Webhook(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		h.shortenURL(w, r)
	} else if r.Method == http.MethodGet {
		h.getFullURL(w, r)
	} else {
		// allow only post/get requests, otherwise send a 405 code
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// POST
func (h *BaseController) shortenURL(w http.ResponseWriter, r *http.Request) {
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
	err = h.storage.Insert(key, string(body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// respond to client
	w.Header().Set("content-type", "text/plain")
	// set code 201
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shurl))
}

// GET
func (h *BaseController) getFullURL(w http.ResponseWriter, r *http.Request) {
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
