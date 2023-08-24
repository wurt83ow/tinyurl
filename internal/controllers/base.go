package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/wurt83ow/tinyurl/cmd/shortener/shorturl"
	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Storage interface {
	Insert(k string, v string) error
	Get(k string) (string, error)
}

type Options interface {
	ParseFlags()
	RunAddr() string
	ShortURLAdress() string
}

type Log interface {
	Info(string, ...zapcore.Field)
}

type BaseController struct {
	storage        Storage
	options        Options
	log            Log
	requestLogger  func(h http.HandlerFunc) http.HandlerFunc
	gzipMiddleware func(h http.HandlerFunc) http.HandlerFunc
}

func NewBaseController(storage Storage, options Options, log Log,
	requestLogger func(h http.HandlerFunc) http.HandlerFunc,
	gzipMiddleware func(h http.HandlerFunc) http.HandlerFunc) *BaseController {

	return &BaseController{storage: storage, options: options, log: log, requestLogger: requestLogger, gzipMiddleware: gzipMiddleware}
}

func (h *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/api/shorten", h.requestLogger(h.gzipMiddleware(h.shortenJSON)))
	r.Post("/", h.requestLogger(h.gzipMiddleware(h.shortenURL)))
	r.Get("/{name}", h.requestLogger(h.gzipMiddleware(h.getFullURL)))
	return r
}

// POST JSON
func (h *BaseController) shortenJSON(w http.ResponseWriter, r *http.Request) {
	// десериализуем запрос в структуру модели
	h.log.Info("decoding request")
	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.log.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.log.Info("request JSON body is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURLAdress := h.options.ShortURLAdress()

	// get short url
	key, shurl := shorturl.Shorten(string(req.URL), shortURLAdress)

	// save full url to storage with the key received earlier
	err := h.storage.Insert(key, string(req.URL))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// заполняем модель ответа
	resp := models.Response{
		Result: shurl,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// сериализуем ответ сервера
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		h.log.Info("error encoding response", zap.Error(err))
		return
	}
	h.log.Info("sending HTTP 201 response")
}

// POST
func (h *BaseController) shortenURL(w http.ResponseWriter, r *http.Request) {

	// установим правильный заголовок для типа данных
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Info("got bad request status 400", zap.String("method", r.Method))
		return
	}
	w.Header().Set("Content-Type", "text/plain")

	shortURLAdress := h.options.ShortURLAdress()

	// get short url
	key, shurl := shorturl.Shorten(string(body), shortURLAdress)

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
	h.log.Info("sending HTTP 201 response")
}

// GET
func (h *BaseController) getFullURL(w http.ResponseWriter, r *http.Request) {

	key := r.URL.Path
	key = strings.Replace(key, "/", "", -1)
	if len(key) == 0 {
		// passed empty key
		w.WriteHeader(http.StatusBadRequest) // 400
		h.log.Info("got bad request status 400", zap.String("method", r.Method))
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
	h.log.Info("temporary redirect status 307")
}
