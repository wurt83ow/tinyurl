package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"

	"github.com/stretchr/testify/assert"
)

func TestWebhook(t *testing.T) {
	// описываем передаваемое тело
	url := "https://practicum.yandex.ru/"
	requestBody := strings.NewReader(url)
	defaultBody := strings.NewReader("")
	// описываем ожидаемое тело ответа при успешном запросе
	successBody := "http://example.com/nOykhckC3Od"

	defaultPath := "/"
	path := "/nOykhckC3Od"

	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		method       string
		path         string
		expectedCode int
		expectedBody string
		requestBody  *strings.Reader
		location     string
	}{
		{method: http.MethodPost, path: defaultPath, expectedCode: http.StatusCreated, expectedBody: successBody, requestBody: requestBody},
		{method: http.MethodGet, path: path, expectedCode: http.StatusTemporaryRedirect, expectedBody: "", requestBody: defaultBody, location: url},
		{method: http.MethodPut, path: defaultPath, expectedCode: http.StatusMethodNotAllowed, expectedBody: "", requestBody: defaultBody},
		{method: http.MethodDelete, path: defaultPath, expectedCode: http.StatusMethodNotAllowed, expectedBody: "", requestBody: defaultBody},
	}

	memoryStorage := storage.NewMemoryStorage()
	handler := NewBaseController(memoryStorage)
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			r := httptest.NewRequest(tc.method, tc.path, tc.requestBody)
			w := httptest.NewRecorder()

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			handler.shortenURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				// assert.JSONEq помогает сравнить две JSON-строки
				assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}

			assert.Equal(t, tc.location, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым")

			handler.getFullURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				// assert.JSONEq помогает сравнить две JSON-строки
				assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}

			assert.Equal(t, tc.location, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым")
		})
	}
}
