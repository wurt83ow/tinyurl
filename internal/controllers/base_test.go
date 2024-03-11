package controllers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	authz "github.com/wurt83ow/tinyurl/internal/authorization"
	"github.com/wurt83ow/tinyurl/internal/config"
	"github.com/wurt83ow/tinyurl/internal/logger"
	"github.com/wurt83ow/tinyurl/internal/models"
	"github.com/wurt83ow/tinyurl/internal/storage"
	"github.com/wurt83ow/tinyurl/internal/worker"
)

var controller *BaseController

func TestMain(m *testing.M) {
	// Perform one-time initialization, such as creating bdKeeper, before running tests
	setup()

	// Run tests
	exitCode := m.Run()

	// Exit with the test completion code
	os.Exit(exitCode)
}

func setup() {
	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		log.Fatalf("Unable to setup logger: %s\n", err)
	}
	keeperMock := new(MockKeeper)
	// Set up a mock for the Load method
	keeperMock.On("Load").Return(storage.StorageURL{}, nil)

	// Set up a mock for the LoadUsers method
	keeperMock.On("LoadUsers").Return(storage.StorageUser{}, nil)

	// Set up a wait for the GetUser method to return an error indicating
	//that the user already exists
	keeperMock.On("GetUser", "test@example.com").Return(storage.ErrConflict)

	data := storage.StorageURL{
		"1": {UUID: "", ShortURL: "", OriginalURL: "https://practicum.yandex.ru/"},
		"2": {UUID: "", ShortURL: "", OriginalURL: "https://www.google.ru/"},		
	}

	keeperMock.On("SaveBatch", data).Return(nil)
	// Set up expectations for methods that will be called inside the Register function
	keeperMock.On("GetUser", "test@example.com").Return(nil) // Example: GetUser method returns an error that the user does not exist
	keeperMock.On("InsertUser", "test@example.com", mock.AnythingOfType("models.DataUser")).Return(nil)

	keeperMock.On("Ping").Return(true)

	// Generate a unique key for the test
	key := "nOykhckC3Od"

	// Generate data for the Save method
	dataURL := models.DataURL{
		UUID:        "",
		OriginalURL: "https://practicum.yandex.ru/",
		ShortURL:    "http://localhost:8080/" + key,
	}

	// Set up a mock for the Save method
	keeperMock.On("Save", key, dataURL).Return(dataURL, nil)

	memoryStorage := storage.NewMemoryStorage(keeperMock, nLogger)

	worker := worker.NewWorker(nLogger, memoryStorage)
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)

	controller = NewBaseController(memoryStorage, option, nLogger, worker, authz)
	if controller == nil {
		log.Fatalf("Unable to initialize baseController\n")
	}
}

func TestShortenJSON(t *testing.T) {

	// describe the body being transmitted
	userReq := strings.NewReader(`{         
        "url": "https://practicum.yandex.ru/"
    }`)

	// describe the expected response body for a successful request
	successBody := `{"result":"http://localhost:8080/nOykhckC3Od"}`

	testPostReq(t, userReq, successBody, "shortenJSON")
}

func TestShortenURL(t *testing.T) {
	// describe the body being transmitted
	url := "https://practicum.yandex.ru/"
	userReq := strings.NewReader(url)

	// describe the expected response body for a successful request
	successBody := "http://localhost:8080/nOykhckC3Od"

	testPostReq(t, userReq, successBody, "shortenURL")
}

func testPostReq(t *testing.T, userReq *strings.Reader, successBody string, funcName string) {

	defaultBody := strings.NewReader("")

	// describe the data set: request method, expected response code, expected body
	testCases := []struct {
		userReq      *strings.Reader
		method       string
		expectedBody string
		expectedCode int
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, expectedBody: successBody, userReq: userReq},
		{method: http.MethodGet, expectedCode: http.StatusBadRequest, expectedBody: "", userReq: defaultBody},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest, expectedBody: "", userReq: defaultBody},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest, expectedBody: "", userReq: defaultBody},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			r := httptest.NewRequest(tc.method, "/", tc.userReq)
			w := httptest.NewRecorder()

			// call the handler as a regular function, without starting the server itself
			switch funcName {
			case "shortenJSON":
				controller.shortenJSON(w, r)
			case "shortenBatch":
				controller.shortenBatch(w, r)
			case "shortenURL":
				controller.shortenURL(w, r)
			}

			assert.Equal(t, tc.expectedCode, w.Code, "The response code does not match what is expected")
			// check the correctness of the received response body if we expect it
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, strings.TrimSpace(w.Body.String()), "The response body does not match what is expected")
			}

		})
	}
}

func TestGetFullURL(t *testing.T) {
	// describe the body being transmitted
	url := "https://practicum.yandex.ru/"

	defaultPath := "/"
	path := "/nOykhckC3Od"

	// describe the data set: request method, expected response code, expected body
	testCases := []struct {
		method       string
		path         string
		location     string
		expectedCode int
	}{
		{method: http.MethodGet, path: path, expectedCode: http.StatusTemporaryRedirect, location: url},
		{method: http.MethodPost, path: defaultPath, expectedCode: http.StatusBadRequest},
		{method: http.MethodPut, path: defaultPath, expectedCode: http.StatusBadRequest},
		{method: http.MethodDelete, path: defaultPath, expectedCode: http.StatusBadRequest},
	}

	// place the data for further retrieval using the get method
	userReq := strings.NewReader(url)
	r := httptest.NewRequest(http.MethodPost, "/", userReq)
	w := httptest.NewRecorder()

	// call the handler as a regular function, without starting the server itself
	controller.shortenURL(w, r)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			r := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			controller.getFullURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "The response code does not match what is expected")

			assert.Equal(t, tc.location, w.Header().Get("Location"), "The Location header is not what you expect")
		})
	}
}

func TestGetPing(t *testing.T) {

	// Create a GET request
	req, err := http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err, "no error was expected when creating the request")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the getPing method
	controller.getPing(rr, req)

	// Check that the status code matches the expected one
	assert.Equal(t, http.StatusOK, rr.Code, "expected status code 200")

	// Check if the code works without the passed keeper
	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		log.Fatalf("Unable to setup logger: %s\n", err)
	}

	memoryStorage := storage.NewMemoryStorage(nil, nLogger)
	worker := worker.NewWorker(nLogger, memoryStorage)
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)

	contr := NewBaseController(memoryStorage, option, nLogger, worker, authz)

	// Create a GET request
	req, err = http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err, "no error was expected when creating the request")

	// Create a ResponseRecorder to record the response
	rr = httptest.NewRecorder()

	// Call the getPing method
	contr.getPing(rr, req)

	// Check that the status code matches the expected one
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "expected status code 500")

}

func TestGetFullURLNotFound(t *testing.T) {
	// describe the data set: request method, expected response code, expected body
	testCases := []struct {
		method       string
		path         string
		expectedCode int
	}{
		{method: http.MethodGet, path: "/nonexistent", expectedCode: http.StatusBadRequest},
		{method: http.MethodPost, path: "/", expectedCode: http.StatusBadRequest},
		{method: http.MethodPut, path: "/", expectedCode: http.StatusBadRequest},
		{method: http.MethodDelete, path: "/", expectedCode: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			controller.getFullURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "The response code does not match what is expected")
		})
	}
}

func TestGetPingNoKeeper(t *testing.T) {
	// Create a new controller without a keeper
	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		log.Fatalf("Unable to setup logger: %s\n", err)
	}

	memoryStorage := storage.NewMemoryStorage(nil, nLogger) // pass nil instead of mock

	worker := worker.NewWorker(nLogger, memoryStorage)
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)

	contr := NewBaseController(memoryStorage, option, nLogger, worker, authz)

	// Create a GET request
	req, err := http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err, "no error was expected when creating the request")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the getPing method
	contr.getPing(rr, req)

	// Check that the status code matches the expected one
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "expected status code 500")
}
