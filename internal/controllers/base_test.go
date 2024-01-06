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
	// Выполните однократную инициализацию, например, создание bdKeeper, перед запуском тестов
	setup()

	// Запустите тесты
	exitCode := m.Run()

	// Выход с кодом завершения тестов
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
	// Настройка мока для метода Load
	keeperMock.On("Load").Return(storage.StorageURL{}, nil)

	// Настройка мока для метода Load
	keeperMock.On("LoadUsers").Return(storage.StorageUser{}, nil)

	// Настраиваем ожидание для метода GetUser, чтобы возвращать ошибку, указывающую, что пользователь уже существует
	keeperMock.On("GetUser", "test@example.com").Return(storage.ErrConflict)

	data := storage.StorageURL{
		"1": {UUID: "", ShortURL: "", OriginalURL: "https://practicum.yandex.ru/"},
		"2": {UUID: "", ShortURL: "", OriginalURL: "https://www.google.ru/"},
		// Add more test data as needed
	}

	keeperMock.On("SaveBatch", data).Return(nil)
	// Настраиваем ожидания для методов, которые будут вызваны внутри функции Register
	keeperMock.On("GetUser", "test@example.com").Return(nil) // Пример: метод GetUser возвращает ошибку, что пользователя не существует
	keeperMock.On("InsertUser", "test@example.com", mock.AnythingOfType("models.DataUser")).Return(nil)

	keeperMock.On("Ping").Return(true)

	// Генерация уникального ключа для теста
	key := "nOykhckC3Od"

	// Генерация данных для метода Save
	dataURL := models.DataURL{
		UUID:        "",
		OriginalURL: "https://practicum.yandex.ru/",
		ShortURL:    "http://localhost:8080/" + key,
	}

	// Настройка мока для метода Save
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

	// Создаем запрос GET
	req, err := http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err, "не ожидалось ошибки при создании запроса")

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем метод getPing
	controller.getPing(rr, req)

	// Проверяем, что статус код соответствует ожидаемому
	assert.Equal(t, http.StatusOK, rr.Code, "ожидался статус код 200")

	// Проверяем работу кода, без переданного keeper
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

	// Создаем запрос GET
	req, err = http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err, "не ожидалось ошибки при создании запроса")

	// Создаем ResponseRecorder для записи ответа
	rr = httptest.NewRecorder()

	// Вызываем метод getPing
	contr.getPing(rr, req)

	// Проверяем, что статус код соответствует ожидаемому
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "ожидался статус код 500")

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
	// Создаем новый контроллер без keeper'а
	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		log.Fatalf("Unable to setup logger: %s\n", err)
	}

	memoryStorage := storage.NewMemoryStorage(nil, nLogger) // передаем nil вместо мока

	worker := worker.NewWorker(nLogger, memoryStorage)
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)

	contr := NewBaseController(memoryStorage, option, nLogger, worker, authz)

	// Создаем запрос GET
	req, err := http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err, "не ожидалось ошибки при создании запроса")

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем метод getPing
	contr.getPing(rr, req)

	// Проверяем, что статус код соответствует ожидаемому
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "ожидался статус код 500")
}
