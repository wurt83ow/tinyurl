package controllers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	authz "github.com/wurt83ow/tinyurl/internal/authorization"
	"github.com/wurt83ow/tinyurl/internal/bdkeeper"
	"github.com/wurt83ow/tinyurl/internal/config"
	"github.com/wurt83ow/tinyurl/internal/filekeeper"
	"github.com/wurt83ow/tinyurl/internal/logger"
	"github.com/wurt83ow/tinyurl/internal/storage"
	"github.com/wurt83ow/tinyurl/internal/worker"
)

func TestShortenJSON(t *testing.T) {

	// describe the body being transmitted
	userReq := strings.NewReader(`{         
        "url": "https://practicum.yandex.ru/"
    }`)

	// describe the expected response body for a successful request
	successBody := `{"result":"http://localhost:8080/nOykhckC3Od"}`

	testPostReq(t, userReq, successBody, "shortenJSON")
}

func TestShortenBatch(t *testing.T) {

	// describe the body being transmitted
	userReq := strings.NewReader(`[
		{
			"correlation_id": "1",
			"original_url": "https://practicum.yandex.ru/"
		},
		 {
			"correlation_id": "2",
			"original_url": "https://www.google.ru/"
		}
	]`)

	// describe the expected response body for a successful request
	successBody := `[
		{
			"correlation_id":"1",
			"short_url":"http://localhost:8080/nOykhckC3Od",
			"original_url":""
		},
		{
			"correlation_id":"2",
			"short_url":"http://localhost:8080/5i80Tt3Jodo",
			"original_url":""
		}
	]`

	successBody = strings.ReplaceAll(successBody, "\n", "")
	successBody = strings.ReplaceAll(successBody, "\t", "")
	testPostReq(t, userReq, strings.TrimSpace(successBody), "shortenBatch")
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
		method       string
		expectedCode int
		expectedBody string
		userReq      *strings.Reader
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, expectedBody: successBody, userReq: userReq},
		{method: http.MethodGet, expectedCode: http.StatusBadRequest, expectedBody: "", userReq: defaultBody},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest, expectedBody: "", userReq: defaultBody},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest, expectedBody: "", userReq: defaultBody},
	}

	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		log.Fatalf("Unable to setup logger: %s\n", err)
	}

	bdKeeper := bdkeeper.NewBDKeeper(option.DataBaseDSN, nLogger)
	var keeper storage.Keeper = nil
	if bdKeeper != nil {
		keeper = bdKeeper
	} else if fileKeeper := filekeeper.NewFileKeeper(option.FileStoragePath, nLogger); fileKeeper != nil {
		keeper = fileKeeper
	}

	if keeper != nil {
		defer keeper.Close()
	}

	memoryStorage := storage.NewMemoryStorage(keeper, nLogger)

	worker := worker.NewWorker(nLogger, memoryStorage)
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)
	controller := NewBaseController(memoryStorage, option, nLogger, worker, authz)

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
		expectedCode int
		location     string
	}{
		{method: http.MethodGet, path: path, expectedCode: http.StatusTemporaryRedirect, location: url},
		{method: http.MethodPost, path: defaultPath, expectedCode: http.StatusBadRequest},
		{method: http.MethodPut, path: defaultPath, expectedCode: http.StatusBadRequest},
		{method: http.MethodDelete, path: defaultPath, expectedCode: http.StatusBadRequest},
	}

	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		log.Fatalf("Unable to setup logger: %s\n", err)
	}

	bdKeeper := bdkeeper.NewBDKeeper(option.DataBaseDSN, nLogger)
	var keeper storage.Keeper = nil
	if bdKeeper != nil {
		keeper = bdKeeper
	} else if fileKeeper := filekeeper.NewFileKeeper(option.FileStoragePath, nLogger); fileKeeper != nil {
		keeper = fileKeeper
	}

	if keeper != nil {
		defer keeper.Close()
	}

	memoryStorage := storage.NewMemoryStorage(keeper, nLogger)

	worker := worker.NewWorker(nLogger, memoryStorage)
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)
	controller := NewBaseController(memoryStorage, option, nLogger, worker, authz)

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
	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		log.Fatalf("Unable to setup logger: %s\n", err)
	}

	bdKeeper := bdkeeper.NewBDKeeper(option.DataBaseDSN, nLogger)
	var keeper storage.Keeper = nil
	if bdKeeper != nil {
		keeper = bdKeeper
	} else if fileKeeper := filekeeper.NewFileKeeper(option.FileStoragePath, nLogger); fileKeeper != nil {
		keeper = fileKeeper
	}

	if keeper != nil {
		defer keeper.Close()
	}

	memoryStorage := storage.NewMemoryStorage(keeper, nLogger)

	worker := worker.NewWorker(nLogger, memoryStorage)
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)
	controller := NewBaseController(memoryStorage, option, nLogger, worker, authz)

	// Создаем запрос GET
	req, err := http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err, "не ожидалось ошибки при создании запроса")

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем метод getPing
	controller.getPing(rr, req)

	// Проверяем, что статус код соответствует ожидаемому
	assert.Equal(t, http.StatusOK, rr.Code, "ожидался статус код 200")

	memoryStorage = storage.NewMemoryStorage(nil, nLogger)
	controller = NewBaseController(memoryStorage, option, nLogger, worker, authz)

	// Создаем запрос GET
	req, err = http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err, "не ожидалось ошибки при создании запроса")

	// Создаем ResponseRecorder для записи ответа
	rr = httptest.NewRecorder()

	// Вызываем метод getPing
	controller.getPing(rr, req)

	// Проверяем, что статус код соответствует ожидаемому
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "ожидался статус код 500")

}
