// Package controllers provides a basic controller for handling user requests.
// It contains the BaseController struct, constructor NewBaseController and its methods for
// authorization, logging and processing user requests.
package controllers

// test remove
import (
	"bytes"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	authz "github.com/wurt83ow/tinyurl/internal/authorization"
	"github.com/wurt83ow/tinyurl/internal/models"
	"github.com/wurt83ow/tinyurl/internal/services/shorturl"
	"github.com/wurt83ow/tinyurl/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var keyUserID models.Key = "userID"

// Storage represents an interface for data storage operations.
type Storage interface {
	// InsertURL inserts a URL entry into the storage.
	InsertURL(k string, v models.DataURL) (models.DataURL, error)

	// InsertUser inserts a user entry into the storage.
	InsertUser(k string, v models.DataUser) (models.DataUser, error)

	// InsertBatch inserts a batch of URL entries into the storage.
	InsertBatch(storageURL storage.StorageURL) error

	// GetURL retrieves a URL entry from the storage.
	GetURL(k string) (models.DataURL, error)

	// GetUser retrieves a user entry from the storage.
	GetUser(k string) (models.DataUser, error)

	// GetUserURLs retrieves URLs associated with a user from the storage.
	GetUserURLs(userID string) []models.DataURLite

	// SaveURL saves a URL entry in the storage.
	SaveURL(k string, v models.DataURL) (models.DataURL, error)

	// DeleteURLs deletes specified URL entries from the storage.
	DeleteURLs(delUrls ...models.DeleteURL) error

	// SaveUser saves a user entry in the storage.
	SaveUser(k string, v models.DataUser) (models.DataUser, error)

	// SaveBatch saves a batch of URL entries in the storage.
	SaveBatch(storageURL storage.StorageURL) error

	// GetBaseConnection checks the base connection status.
	GetBaseConnection() bool
}

// Options represents an interface for parsing command line options.
type Options interface {
	// ParseFlags parses command line flags.
	ParseFlags()

	// RunAddr returns the address to run the application.
	RunAddr() string

	// ShortURLAdress returns the short URL address.
	ShortURLAdress() string
}

// Log represents an interface for logging functionality.
type Log interface {
	// Info logs an informational message with optional fields.
	Info(string, ...zapcore.Field)
}

// Worker represents an interface for worker functionality.
type Worker interface {
	// Add adds a task to the worker.
	Add(models.DeleteURL)
}

// Authz represents an interface for user authorization functionality.
type Authz interface {
	// JWTAuthzMiddleware returns a middleware function for JWT-based authorization.
	JWTAuthzMiddleware(authz.Storage, authz.Log) func(http.Handler) http.Handler

	// GetHash generates a hash for a given email and password.
	GetHash(email string, password string) []byte

	// CreateJWTTokenForUser creates a JWT token for a specified user ID.
	CreateJWTTokenForUser(userID string) string

	// AuthCookie creates an HTTP cookie for authorization purposes.
	AuthCookie(name string, token string) *http.Cookie
}

// BaseController represents a basic controller for handling user requests.
// It includes handler methods for various operations.
type BaseController struct {
	storage Storage
	options Options
	log     Log
	worker  Worker
	authz   Authz
}

// Example usage:
//
//	controller := NewBaseController(memoryStorage, option, nLogger, worker, authz)
//	r.Mount("/", controller.Route())
//	flagRunAddr := option.RunAddr()
//	http.ListenAndServe(flagRunAddr, r)
func NewBaseController(storage Storage, options Options, log Log, worker Worker, authz Authz) *BaseController {
	instance := &BaseController{
		storage: storage,
		options: options,
		log:     log,
		worker:  worker,
		authz:   authz,
		// delChan: make(chan models.DeleteURL, 1024), // set the channel buffer to 1024 messages
	}

	// go instance.flushURLs()

	return instance
}

// Route returns a chi.Mux router with registered handlers for BaseController routes.
// It creates a new chi router, registers the routes with the corresponding handler
// methods, and returns the configured router.
//
// Receiver:
//   - h: A pointer to the BaseController instance.
//
// Returns:
//
//	A chi.Mux router with registered routes and handlers.
func (h *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Get("/{name}", h.getFullURL)
	r.Get("/ping", h.getPing)

	r.Get("/pprof/*", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)
	r.Handle("/vars", expvar.Handler())

	r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/pprof/mutex", pprof.Handler("mutex"))
	r.Handle("/pprof/heap", pprof.Handler("heap"))
	r.Handle("/pprof/block", pprof.Handler("block"))
	r.Handle("/pprof/allocs", pprof.Handler("allocs"))

	// group where the middleware authorization is needed
	r.Group(func(r chi.Router) {
		r.Use(h.authz.JWTAuthzMiddleware(h.storage, h.log))

		r.Post("/", h.shortenURL)
		r.Post("/api/shorten", h.shortenJSON)
		r.Post("/api/shorten/batch", h.shortenBatch)
		r.Get("/api/user/urls", h.getUserURLs)
		r.Delete("/api/user/urls", h.deleteUserURLs)
	})

	return r
}

// deleteUserURLs is a handler method for deleting user URLs.
// It takes a pointer to the BaseController instance, an http.ResponseWriter, and an http.Request as parameters.
// The function decodes the request JSON body containing URL IDs to be deleted, validates the user ID from the request context,
// adds a task to the worker for asynchronous deletion, and responds with the appropriate HTTP status code.
//
// Parameters:
//   - h: A pointer to the BaseController instance.
//   - w: An http.ResponseWriter for writing the HTTP response.
//   - r: An http.Request representing the incoming HTTP request.
func (h *BaseController) deleteUserURLs(w http.ResponseWriter, r *http.Request) {
	// Initialize an empty slice to store URL IDs
	ids := make([]string, 0)

	// Create a JSON decoder for decoding the request body
	dec := json.NewDecoder(r.Body)

	// Decode the request JSON body into the ids slice
	if err := dec.Decode(&ids); err != nil {
		// Log the error and respond with a Bad Request status code
		h.log.Info("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Retrieve the user ID from the request context
	userID, ok := r.Context().Value(keyUserID).(string)

	// Check if user ID is present in the context
	if !ok {
		// Respond with an Unauthorized status code
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Add a task to the worker for asynchronous deletion
	h.worker.Add(models.DeleteURL{UserID: userID, ShortURLs: ids})

	// Respond with an Accepted status code
	w.WriteHeader(http.StatusAccepted)
}

// Register is a method of the *BaseController structure.
//
// The method handles user registration by accepting the email and password
// as parameters in the request body. Upon successful registration (user creation),
// it returns a HTTP status code 201. If the user already exists in the system
// (previously created), it returns a HTTP status code 409. In case of invalid
// email or password, it returns a HTTP status code 400.
//
// # Example
//
// Code:
//
//	func (h *BaseController) Route() *chi.Mux {
//		r := chi.NewRouter()
//		r.Post("/register", h.Register)
//
//	  return r
//	}
func (h *BaseController) Register(w http.ResponseWriter, r *http.Request) {
	regReq := models.RequestUser{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&regReq); err != nil {
		h.log.Info("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := h.storage.GetUser(regReq.Email)
	fmt.Println(regReq.Email)
	if err == nil {
		h.log.Info("the user is already registered: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	Hash := h.authz.GetHash(regReq.Email, regReq.Password)

	// save the user to the storage
	dataUser := models.DataUser{UUID: uuid.New().String(), Email: regReq.Email, Hash: Hash, Name: regReq.Name}

	_, err = h.storage.InsertUser(regReq.Email, dataUser)

	if err != nil {
		if err == storage.ErrConflict {
			w.WriteHeader(http.StatusConflict) // code 409
		} else {
			w.WriteHeader(http.StatusBadRequest) // code 400
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	h.log.Info("sending HTTP 201 response")
}

// Login is a handler method for user authentication and login.
// It takes a pointer to the BaseController instance, an http.ResponseWriter, and an http.Request as parameters.
// The function decodes the request body into a RequestUser model, retrieves the user from storage based on the provided email,
// validates the password, generates and sets a new JWT token in cookies, and responds with the appropriate HTTP status code.
//
// Parameters:
//   - h: A pointer to the BaseController instance.
//   - w: An http.ResponseWriter for writing the HTTP response.
//   - r: An http.Request representing the incoming HTTP request.
//
// # Example
//
// Code:
//
//	   func (h *BaseController) Route() *chi.Mux {
//	       r := chi.NewRouter()
//			  r.Post("/login", h.Login)
//
//		      return r
//	}
//
// Output: Login successful, Status Code: 200
func (h *BaseController) Login(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into a RequestUser model
	var rb models.RequestUser
	if err := json.NewDecoder(r.Body).Decode(&rb); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Retrieve the user from storage based on the provided email
	user, err := h.storage.GetUser(rb.Email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate the password
	if bytes.Equal(user.Hash, h.authz.GetHash(rb.Email, rb.Password)) {
		// Generate a new JWT token for the user
		freshToken := h.authz.CreateJWTTokenForUser(user.UUID)

		// Set the JWT token in cookies and response headers
		http.SetCookie(w, h.authz.AuthCookie("jwt-token", freshToken))
		http.SetCookie(w, h.authz.AuthCookie("Authorization", freshToken))
		w.Header().Set("Authorization", freshToken)

		// Respond with a success message
		err = json.NewEncoder(w).Encode(models.ResponseUser{
			Response: "success",
		})

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		return
	}

	// Respond with an incorrect email/password message
	err = json.NewEncoder(w).Encode(models.ResponseUser{
		Response: "incorrect email/password",
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// shortenBatch is a handler method for batch URL shortening.
// It takes a pointer to the BaseController instance, an http.ResponseWriter, and an http.Request as parameters.
// The function deserializes the request JSON body into a model structure, performs URL shortening for each URL in the batch,
// saves the shortened URLs to storage, and responds with the appropriate HTTP status code and serialized response.
//
// Parameters:
//   - h: A pointer to the BaseController instance.
//   - w: An http.ResponseWriter for writing the HTTP response.
//   - r: An http.Request representing the incoming HTTP request.
func (h *BaseController) shortenBatch(w http.ResponseWriter, r *http.Request) {
	// Deserialize the request into the model structure
	h.log.Info("decoding request")

	// Initialize a batch slice to store DataURLite models
	batch := []models.DataURLite{}

	// Create a JSON decoder for decoding the request body
	dec := json.NewDecoder(r.Body)

	// Decode the request JSON body into the batch slice
	if err := dec.Decode(&batch); err != nil {
		// Log the error and respond with a Bad Request status code
		h.log.Info("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the request JSON body is empty
	if len(batch) == 0 {
		// Log the error and respond with a Bad Request status code
		h.log.Info("request JSON body is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the short URL address from the options
	shortURLAdress := h.options.ShortURLAdress()

	// Initialize a storageURL map to store data for batch insertion
	dataURL := make(storage.StorageURL)

	// Initialize a response slice to store the shortened URLs for the client response
	resp := []models.DataURLite{}

	// Retrieve the user ID from the request context
	userID, _ := r.Context().Value(keyUserID).(string)

	// Loop through each URL in the batch
	for i := range batch {
		s := batch[i]

		// Shorten the original URL
		key, shurl := shorturl.Shorten(s.OriginalURL, shortURLAdress)

		// Save the full URL to storage with the key received earlier
		data := models.DataURL{UUID: s.UUID, ShortURL: shurl, OriginalURL: s.OriginalURL, UserID: userID}
		dataURL[key] = data

		// Append the shortened URL to the response slice
		resp = append(resp, models.DataURLite{UUID: s.UUID, ShortURL: shurl})
	}

	// Insert the batch of URLs into the storage
	err := h.storage.InsertBatch(dataURL)
	if err != nil {
		// Respond with a Bad Request status code if there is an error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Serialize the server response
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		// Log the error if there is an encoding issue
		h.log.Info("error encoding response: ", zap.Error(err))
		return
	}

	// Log the successful HTTP 201 response
	h.log.Info("sending HTTP 201 response")
}

// shortenJSON is a handler method for shortening a single URL from a JSON request.
// It takes a pointer to the BaseController instance, an http.ResponseWriter, and an http.Request as parameters.
// The function deserializes the request JSON body into a Request model, shortens the URL, saves it to storage,
// and responds with the appropriate HTTP status code and serialized response.
//
// Parameters:
//   - h: A pointer to the BaseController instance.
//   - w: An http.ResponseWriter for writing the HTTP response.
//   - r: An http.Request representing the incoming HTTP request.
func (h *BaseController) shortenJSON(w http.ResponseWriter, r *http.Request) {
	// Deserialize the request into the model structure
	h.log.Info("decoding request")

	// Initialize a Request model
	var req models.Request

	// Create a JSON decoder for decoding the request body
	dec := json.NewDecoder(r.Body)

	// Decode the request JSON body into the Request model
	if err := dec.Decode(&req); err != nil {
		// Log the error and respond with a Bad Request status code
		h.log.Info("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the request JSON body is empty
	if req.URL == "" {
		// Log the error and respond with a Bad Request status code
		h.log.Info("request JSON body is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the short URL address from the options
	shortURLAdress := h.options.ShortURLAdress()

	// Shorten the original URL
	key, shurl := shorturl.Shorten(string(req.URL), shortURLAdress)

	// Retrieve the user ID from the request context
	userID, _ := r.Context().Value(keyUserID).(string)

	// Save the full URL to storage with the key received earlier
	m, err := h.storage.InsertURL(key, models.DataURL{ShortURL: shurl, OriginalURL: string(req.URL), UserID: userID})

	// Check for conflicts or other errors during insertion
	conflict := false
	if err != nil {
		if err == storage.ErrConflict {
			conflict = true
		} else {
			// Respond with a Bad Request status code for other errors
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Fill in the response model
	resp := models.Response{
		Result: m.ShortURL,
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	// Serialize the server response
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		// Log the error if there is an encoding issue
		h.log.Info("error encoding response: ", zap.Error(err))
		return
	}

	// Log the successful HTTP 201 response
	h.log.Info("sending HTTP 201 response")
}

// shortenURL is a handler method for shortening a single URL.
// It takes a pointer to the BaseController instance, an http.ResponseWriter, and an http.Request as parameters.
// The function reads the request body, shortens the URL, saves it to storage, and responds with the appropriate HTTP status code.
//
// Parameters:
//   - h: A pointer to the BaseController instance.
//   - w: An http.ResponseWriter for writing the HTTP response.
//   - r: An http.Request representing the incoming HTTP request.
func (h *BaseController) shortenURL(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		// Respond with a Bad Request status code if there is an error or the body is empty
		w.WriteHeader(http.StatusBadRequest)
		h.log.Info("got bad request status 400", zap.String("method", r.Method))
		return
	}

	// Set the correct content type header
	w.Header().Set("Content-Type", "text/plain")

	// Get the short URL address from the options
	shortURLAdress := h.options.ShortURLAdress()

	// Shorten the URL
	key, shurl := shorturl.Shorten(string(body), shortURLAdress)

	// Retrieve the user ID from the request context
	userID, _ := r.Context().Value(keyUserID).(string)

	// Save the full URL to storage with the key received earlier
	m, err := h.storage.InsertURL(key, models.DataURL{ShortURL: shurl, OriginalURL: string(body), UserID: userID})

	// Check for conflicts or other errors during insertion
	conflict := false
	if err != nil {
		if err == storage.ErrConflict {
			conflict = true
		} else {
			// Respond with a Bad Request status code for other errors
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Respond to the client
	if conflict {
		w.WriteHeader(http.StatusConflict) // Code 409
	} else {
		w.WriteHeader(http.StatusCreated) // Code 201
	}

	// Write the shortened URL to the response
	_, err = w.Write([]byte(m.ShortURL))
	if err != nil {
		// Respond with a Bad Request status code if there is an error writing the response
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Log the successful HTTP 201 response
	h.log.Info("sending HTTP 201 response")
}

// getFullURL is a handler method for retrieving the original URL for a given shortened URL key.
// It takes a pointer to the BaseController instance, an http.ResponseWriter, and an http.Request as parameters.
// The function processes a custom GET request, retrieves the original URL from storage, and responds with the appropriate HTTP status code.
//
// Parameters:
//   - h: A pointer to the BaseController instance.
//   - w: An http.ResponseWriter for writing the HTTP response.
//   - r: An http.Request representing the incoming HTTP request.
func (h *BaseController) getFullURL(w http.ResponseWriter, r *http.Request) {
	// Extract the key from the URL path
	key := r.URL.Path
	key = strings.Replace(key, "/", "", -1)

	// Respond with a Bad Request status code if the key is empty
	if len(key) == 0 {
		w.WriteHeader(http.StatusBadRequest) // Code 400
		h.log.Info("got bad request status 400", zap.String("method", r.Method))
		return
	}

	// Get the full URL from storage
	data, err := h.storage.GetURL(key)

	// Respond with a Bad Request status code if the URL is not found or there is an error
	if err != nil || len(data.OriginalURL) == 0 {
		w.WriteHeader(http.StatusBadRequest) // Code 400
		return
	}

	// Respond with a Gone status code if the URL has been marked as deleted
	if data.DeletedFlag {
		w.WriteHeader(http.StatusGone) // Code 410
		return
	}

	// Set the Location header for a temporary redirect
	w.Header().Set("Location", data.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect) // Code 307
	h.log.Info("temporary redirect status 307")
}

// getUserURLs is a handler method for retrieving URLs associated with the authenticated user.
// It takes a pointer to the BaseController instance, an http.ResponseWriter, and an http.Request as parameters.
// The function retrieves the user ID from the request context, retrieves associated URLs from storage,
// and responds with the appropriate HTTP status code and serialized response.
//
// Parameters:
//   - h: A pointer to the BaseController instance.
//   - w: An http.ResponseWriter for writing the HTTP response.
//   - r: An http.Request representing the incoming HTTP request.
func (h *BaseController) getUserURLs(w http.ResponseWriter, r *http.Request) {
	// Retrieve the user ID from the request context
	userID, ok := r.Context().Value(keyUserID).(string)
	if !ok {
		// Respond with an Unauthorized status code if the user ID is not present
		w.WriteHeader(http.StatusUnauthorized) // Code 401
		return
	}

	// Retrieve URLs associated with the user from storage
	data := h.storage.GetUserURLs(userID)

	// Respond with a No Content status code if no URLs are found for the user
	if len(data) == 0 {
		w.WriteHeader(http.StatusNoContent) // Code 204
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Respond with an OK status code and serialize the response
	w.WriteHeader(http.StatusOK) // Code 200
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		h.log.Info("error encoding response: ", zap.Error(err))
		return
	}
}

// getPing is a handler method for processing incoming GET requests and sending a response based on storage availability.
// It takes a pointer to the BaseController instance, an http.ResponseWriter, and an http.Request as parameters.
// The function responds with an OK status code if the storage is available, or an Internal Server Error status code if it is not.
//
// Parameters:
//   - h: A pointer to the BaseController instance.
//   - w: An http.ResponseWriter for writing the HTTP response.
//   - r: An http.Request representing the incoming HTTP request.
func (h *BaseController) getPing(w http.ResponseWriter, r *http.Request) {
	// Check if the storage (database or file JSON) is available
	if !h.storage.GetBaseConnection() {
		// Respond with an Internal Server Error status code if the storage
		h.log.Info("got status internal server error")
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}

	w.WriteHeader(http.StatusOK) // 200
	h.log.Info("sending HTTP 200 response")
}
