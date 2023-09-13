package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/wurt83ow/tinyurl/cmd/shortener/shorturl"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"
	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Storage interface {
	Insert(k string, v models.DataURL) (models.DataURL, error)
	InsertBatch(storage.StorageURL) error
	Get(k string) (models.DataURL, error)
	Save(k string, v models.DataURL) (models.DataURL, error)
	SaveBatch(storage.StorageURL) error
	GetBaseConnection() bool
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
	storage Storage
	options Options
	log     Log
}

func NewBaseController(storage Storage, options Options, log Log) *BaseController {

	return &BaseController{storage: storage, options: options, log: log}
}

func (h *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/api/shorten", h.shortenJSON)
	r.Post("/api/shorten/batch", h.shortenBatch)
	r.Post("/", h.shortenURL)
	r.Get("/{name}", h.getFullURL)
	r.Get("/ping", h.getPing)
	return r
}

// POST JSON BATCH
func (h *BaseController) shortenBatch(w http.ResponseWriter, r *http.Request) {
	// десериализуем запрос в структуру модели
	h.log.Info("decoding request")

	batch := []models.RequestRecord{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&batch); err != nil {
		h.log.Info("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(batch) == 0 {
		h.log.Info("request JSON body is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURLAdress := h.options.ShortURLAdress()
	dataURL := make(storage.StorageURL)
	resp := []models.ResponseRecord{}
	for i := range batch {

		s := batch[i]
		// get short url
		key, shurl := shorturl.Shorten(s.OriginalURL, shortURLAdress)

		// save full url to storage with the key received earlier
		data := models.DataURL{UUID: s.UUID, ShortURL: shurl, OriginalURL: s.OriginalURL}
		dataURL[key] = data
		resp = append(resp, models.ResponseRecord{UUID: s.UUID, ShortURL: shurl})

	}

	err := h.storage.InsertBatch(dataURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// заполняем модель ответа

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// сериализуем ответ сервера
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		h.log.Info("error encoding response: ", zap.Error(err))
		return
	}
	h.log.Info("sending HTTP 201 response")
}

// POST JSON
func (h *BaseController) shortenJSON(w http.ResponseWriter, r *http.Request) {
	// десериализуем запрос в структуру модели
	h.log.Info("decoding request")
	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.log.Info("cannot decode request JSON body: ", zap.Error(err))
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
	m, err := h.storage.Insert(key, models.DataURL{ShortURL: shurl, OriginalURL: string(req.URL)})
	conflict := false

	if err != nil {
		if err == storage.ErrConflict {
			conflict = true
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// заполняем модель ответа
	resp := models.Response{
		Result: m.ShortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	// сериализуем ответ сервера
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		h.log.Info("error encoding response: ", zap.Error(err))
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
		h.log.Info("got bad request status 400: %v", zap.String("method", r.Method))
		return
	}
	w.Header().Set("Content-Type", "text/plain")

	shortURLAdress := h.options.ShortURLAdress()

	// get short url
	key, shurl := shorturl.Shorten(string(body), shortURLAdress)

	// save full url to storage with the key received earlier
	m, err := h.storage.Insert(key, models.DataURL{ShortURL: shurl, OriginalURL: string(body)})
	conflict := false
	if err != nil {
		if err == storage.ErrConflict {
			conflict = true
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	// respond to client
	w.Header().Set("Content-Type", "text/plain")

	if conflict {
		w.WriteHeader(http.StatusConflict) //code 409
	} else {
		w.WriteHeader(http.StatusCreated) //code 201
	}

	w.Write([]byte(m.ShortURL))
	h.log.Info("sending HTTP 201 response")
}

// GET
func (h *BaseController) getFullURL(w http.ResponseWriter, r *http.Request) {

	key := r.URL.Path
	key = strings.Replace(key, "/", "", -1)
	if len(key) == 0 {
		// passed empty key
		w.WriteHeader(http.StatusBadRequest) // 400
		h.log.Info("got bad request status 400: %v", zap.String("method", r.Method))
		return
	}
	// get full url from storage
	data, err := h.storage.Get(key)
	if err != nil || len(data.OriginalURL) == 0 {
		// value not found for the passed key
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	w.Header().Set("Location", data.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect) // 307
	h.log.Info("temporary redirect status 307")
}

// GET
func (h *BaseController) getPing(w http.ResponseWriter, r *http.Request) {

	if !h.storage.GetBaseConnection() {
		h.log.Info("got status internal server error")
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
	w.WriteHeader(http.StatusOK) // 200
	h.log.Info("sending HTTP 200 response")
}
