package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wurt83ow/tinyurl/cmd/shortener/config"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"
	"github.com/wurt83ow/tinyurl/internal/logger"
)

func shortenJSON(t *testing.T) {
	// описываем передаваемое тело

	requestBody := strings.NewReader(`{         
        "url": "https://practicum.yandex.ru/"
    }`)

	defaultBody := strings.NewReader("")
	// описываем ожидаемое тело ответа при успешном запросе
	successBody := `{         
        "result": "http://localhost:8080/nOykhckC3Od"
    }`

	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		method       string
		expectedCode int
		expectedBody string
		requestBody  *strings.Reader
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, expectedBody: successBody, requestBody: requestBody},
		{method: http.MethodGet, expectedCode: http.StatusBadRequest, expectedBody: "", requestBody: defaultBody},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest, expectedBody: "", requestBody: defaultBody},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest, expectedBody: "", requestBody: defaultBody},
	}

	option := config.NewOptions()
	option.ParseFlags()

	if err := logger.Initialize(option.LogLevel()); err != nil {
		return
	}

	memoryStorage := storage.NewMemoryStorage()

	handler := NewBaseController(memoryStorage, option, logger.Log, logger.RequestLogger)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			r := httptest.NewRequest(tc.method, "/", tc.requestBody)
			w := httptest.NewRecorder()

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			handler.shortenURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}

		})
	}
}

func TestShortenURL(t *testing.T) {
	// описываем передаваемое тело
	url := "https://practicum.yandex.ru/"
	requestBody := strings.NewReader(url)
	defaultBody := strings.NewReader("")
	// описываем ожидаемое тело ответа при успешном запросе
	successBody := "http://localhost:8080/nOykhckC3Od"

	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		method       string
		expectedCode int
		expectedBody string
		requestBody  *strings.Reader
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, expectedBody: successBody, requestBody: requestBody},
		{method: http.MethodGet, expectedCode: http.StatusBadRequest, expectedBody: "", requestBody: defaultBody},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest, expectedBody: "", requestBody: defaultBody},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest, expectedBody: "", requestBody: defaultBody},
	}

	option := config.NewOptions()
	option.ParseFlags()

	if err := logger.Initialize(option.LogLevel()); err != nil {
		return
	}

	memoryStorage := storage.NewMemoryStorage()

	handler := NewBaseController(memoryStorage, option, logger.Log, logger.RequestLogger)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			r := httptest.NewRequest(tc.method, "/", tc.requestBody)
			w := httptest.NewRecorder()

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			handler.shortenURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}

		})
	}
}

func TestGetFullURL(t *testing.T) {
	// описываем передаваемое тело
	url := "https://practicum.yandex.ru/"

	defaultPath := "/"
	path := "/nOykhckC3Od"

	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
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

	if err := logger.Initialize(option.LogLevel()); err != nil {
		return
	}
	memoryStorage := storage.NewMemoryStorage()

	handler := NewBaseController(memoryStorage, option, logger.Log, logger.RequestLogger)

	//Поместим данные для дальнейшего их получения методом get
	requestBody := strings.NewReader(url)
	r := httptest.NewRequest(http.MethodPost, "/", requestBody)
	w := httptest.NewRecorder()

	// вызовем хендлер как обычную функцию, без запуска самого сервера
	handler.shortenURL(w, r)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			r := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			handler.getFullURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")

			assert.Equal(t, tc.location, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым")
		})
	}
}
