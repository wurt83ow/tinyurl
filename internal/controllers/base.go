package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	authz "github.com/wurt83ow/tinyurl/cmd/shortener/authorization"
	"github.com/wurt83ow/tinyurl/cmd/shortener/shorturl"
	"github.com/wurt83ow/tinyurl/internal/middleware"
	"github.com/wurt83ow/tinyurl/internal/models"
	"github.com/wurt83ow/tinyurl/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var keyUserID models.Key = "userID"

type Storage interface {
	InsertURL(k string, v models.DataURL) (models.DataURL, error)
	InsertUser(k string, v models.DataUser) (models.DataUser, error)
	InsertBatch(storage.StorageURL) error
	GetURL(k string) (models.DataURL, error)
	GetUser(k string) (models.DataUser, error)
	GetUserURLs(userID string) []models.ResponseUserURLs
	SaveURL(k string, v models.DataURL) (models.DataURL, error)
	DeleteURLs(delUrls ...models.DeleteURL) error
	SaveUser(k string, v models.DataUser) (models.DataUser, error)
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

	delChan chan models.DeleteURL
}

func NewBaseController(storage Storage, options Options, log Log) *BaseController {
	instance := &BaseController{
		storage: storage,
		options: options,
		log:     log,
		delChan: make(chan models.DeleteURL, 1024), // установим каналу буфер в 1024 сообщения
	}

	go instance.flushURLs()

	return instance
}

func (h *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Get("/{name}", h.getFullURL)
	r.Get("/ping", h.getPing)

	// group where the middleware authorization is needed
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuthzMiddleware(h.storage, h.log))

		r.Post("/", h.shortenURL)
		r.Post("/api/shorten", h.shortenJSON)
		r.Post("/api/shorten/batch", h.shortenBatch)
		r.Get("/api/user/urls", h.getUserURLs)
		r.Delete("/api/user/urls", h.deleteUserURLs)
	})

	return r
}

// flushURLs постоянно сохраняет несколько сообщений в хранилище с определённым интервалом
func (h *BaseController) flushURLs() {
	// будем сохранять сообщения, накопленные за последние 10 секунд
	ticker := time.NewTicker(1 * time.Second)
	// ticker := time.NewTicker(time.Millisecond)
	var delUrls []models.DeleteURL
	var mu sync.Mutex
	for {
		select {
		case msg := <-h.delChan:
			// добавим сообщение в слайс для последующего сохранения
			mu.Lock()
			delUrls = append(delUrls, msg)
			mu.Unlock()
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(delUrls) == 0 {
				continue
			}
			// сохраним все пришедшие сообщения одновременно
			err := h.storage.DeleteURLs(delUrls...)
			if err != nil {
				h.log.Info("cannot save delUrls", zap.Error(err))
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			delUrls = nil
		}
	}
}

func (h *BaseController) deleteUserURLs(w http.ResponseWriter, r *http.Request) {

	ids := make([]string, 0)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&ids); err != nil {
		h.log.Info("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	userID, ok := r.Context().Value(keyUserID).(string)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized) //401
		return
	}
	h.delChan <- models.DeleteURL{UserID: userID, ShortURLs: ids}

	w.WriteHeader(http.StatusAccepted)

}

func (h *BaseController) Register(w http.ResponseWriter, r *http.Request) {

	regReq := models.RegisterRequest{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&regReq); err != nil {
		h.log.Info("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := h.storage.GetUser(regReq.Email)
	if err == nil {
		h.log.Info("the user is already registered: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	Hash := authz.GetHash(regReq.Email, regReq.Password)

	// save the user to the storage
	dataUser := models.DataUser{UUID: uuid.New().String(), Email: regReq.Email, Hash: Hash, Name: regReq.Name}

	_, err = h.storage.InsertUser(regReq.Email, dataUser)

	if err != nil {
		if err == storage.ErrConflict {
			w.WriteHeader(http.StatusConflict) //code 409
		} else {
			w.WriteHeader(http.StatusBadRequest) // code 400
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	h.log.Info("sending HTTP 201 response")

}

func (h *BaseController) Login(w http.ResponseWriter, r *http.Request) {

	var rb models.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&rb); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUser(rb.Email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if bytes.Equal(user.Hash, authz.GetHash(rb.Email, rb.Password)) {
		freshToken := authz.CreateJWTTokenForUser(user.UUID)
		http.SetCookie(w, authz.AuthCookie("jwt-token", freshToken))
		http.SetCookie(w, authz.AuthCookie("Authorization", freshToken))

		w.Header().Set("Authorization", freshToken)
		err := json.NewEncoder(w).Encode(models.ResponseBody{
			Response: "success",
		})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		return
	}

	err = json.NewEncoder(w).Encode(models.ResponseBody{
		Response: "incorrect email/password",
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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

	userID, _ := r.Context().Value(keyUserID).(string)

	for i := range batch {

		s := batch[i]
		// get short url
		key, shurl := shorturl.Shorten(s.OriginalURL, shortURLAdress)

		// save full url to storage with the key received earlier
		data := models.DataURL{UUID: s.UUID, ShortURL: shurl, OriginalURL: s.OriginalURL, UserID: userID}
		dataURL[key] = data
		resp = append(resp, models.ResponseRecord{UUID: s.UUID, ShortURL: shurl})
	}

	err := h.storage.InsertBatch(dataURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fill in the response model
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// serialize the server response
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		h.log.Info("error encoding response: ", zap.Error(err))
		return
	}
	h.log.Info("sending HTTP 201 response")
}

// POST JSON
func (h *BaseController) shortenJSON(w http.ResponseWriter, r *http.Request) {

	// deserialize the request into the model structure
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
	userID, _ := r.Context().Value(keyUserID).(string)

	// save full url to storage with the key received earlier
	m, err := h.storage.InsertURL(key, models.DataURL{ShortURL: shurl, OriginalURL: string(req.URL), UserID: userID})
	conflict := false

	if err != nil {
		if err == storage.ErrConflict {
			conflict = true
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// fill in the response model
	resp := models.Response{
		Result: m.ShortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	// serialize the server response
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		h.log.Info("error encoding response: ", zap.Error(err))
		return
	}
	h.log.Info("sending HTTP 201 response")
}

// POST
func (h *BaseController) shortenURL(w http.ResponseWriter, r *http.Request) {

	// set the correct header for the data type
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

	userID, _ := r.Context().Value(keyUserID).(string)

	// save full url to storage with the key received earlier
	m, err := h.storage.InsertURL(key, models.DataURL{ShortURL: shurl, OriginalURL: string(body), UserID: userID})

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

	_, err = w.Write([]byte(m.ShortURL))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
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
	data, err := h.storage.GetURL(key)
	if err != nil || len(data.OriginalURL) == 0 {

		// value not found for the passed key
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	if data.DeletedFlag {
		w.WriteHeader(http.StatusGone) // 410
		return
	}

	w.Header().Set("Location", data.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect) // 307
	h.log.Info("temporary redirect status 307")
}

// GET
func (h *BaseController) getUserURLs(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(keyUserID).(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized) //401
		return
	}

	data := h.storage.GetUserURLs(userID)
	if len(data) == 0 {
		// value not found for the passed key
		w.WriteHeader(http.StatusNoContent) // 204
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) //code 200

	// serialize the server response
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		h.log.Info("error encoding response: ", zap.Error(err))
		return
	}
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
