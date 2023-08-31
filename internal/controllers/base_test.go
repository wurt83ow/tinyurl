package controllers

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wurt83ow/tinyurl/cmd/shortener/config"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"
	"github.com/wurt83ow/tinyurl/internal/compressor"
	"github.com/wurt83ow/tinyurl/internal/keeper"
	"github.com/wurt83ow/tinyurl/internal/logger"
)

func TestShortenJSON(t *testing.T) {

	// описываем передаваемое тело
	requestBody := strings.NewReader(`{         
        "url": "https://practicum.yandex.ru/"
    }`)

	// описываем ожидаемое тело ответа при успешном запросе
	successBody := `{"result":"http://localhost:8080/nOykhckC3Od"}`

	testPostReq(t, requestBody, successBody, true)
}

func TestShortenURL(t *testing.T) {
	// описываем передаваемое тело
	url := "https://practicum.yandex.ru/"
	requestBody := strings.NewReader(url)

	// описываем ожидаемое тело ответа при успешном запросе
	successBody := "http://localhost:8080/nOykhckC3Od"

	testPostReq(t, requestBody, successBody, false)
}

func testPostReq(t *testing.T, requestBody *strings.Reader, successBody string, isJSONTest bool) {

	defaultBody := strings.NewReader("")

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

	keeper := keeper.NewKeeper(option.FileStoragePath, logger.Log)
	memoryStorage := storage.NewMemoryStorage(keeper, logger.Log)

	controller := NewBaseController(memoryStorage, option, logger.Log, logger.RequestLogger, compressor.GzipMiddleware)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			r := httptest.NewRequest(tc.method, "/", tc.requestBody)
			w := httptest.NewRecorder()

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			if isJSONTest {
				controller.shortenJSON(w, r)
			} else {
				controller.shortenURL(w, r)
			}

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, strings.TrimSpace(w.Body.String()), "Тело ответа не совпадает с ожидаемым")
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
	keeper := keeper.NewKeeper(option.FileStoragePath, logger.Log)
	memoryStorage := storage.NewMemoryStorage(keeper, logger.Log)

	controller := NewBaseController(memoryStorage, option, logger.Log, logger.RequestLogger, compressor.GzipMiddleware)

	//Поместим данные для дальнейшего их получения методом get
	requestBody := strings.NewReader(url)
	r := httptest.NewRequest(http.MethodPost, "/", requestBody)
	w := httptest.NewRecorder()

	// вызовем хендлер как обычную функцию, без запуска самого сервера
	controller.shortenURL(w, r)

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			r := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			controller.getFullURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")

			assert.Equal(t, tc.location, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым")
		})
	}
}

func TestGzipShortenJSON(t *testing.T) {

	// описываем передаваемое тело
	requestBody := `{         
        "url": "https://practicum.yandex.ru/"
    }`

	// описываем ожидаемое тело ответа при успешном запросе
	successBody := `{"result":"http://localhost:8080/nOykhckC3Od"}`

	testGzipCompression(t, requestBody, successBody, true)
}

func TestGzipTestShortenURL(t *testing.T) {
	// описываем передаваемое тело
	requestBody := "https://practicum.yandex.ru/"

	// описываем ожидаемое тело ответа при успешном запросе
	successBody := "http://localhost:8080/nOykhckC3Od"

	testGzipCompression(t, requestBody, successBody, false)
}

func testGzipCompression(t *testing.T, requestBody string, successBody string, isJSONTest bool) {

	option := config.NewOptions()
	option.ParseFlags()

	if err := logger.Initialize(option.LogLevel()); err != nil {
		return
	}

	keeper := keeper.NewKeeper(option.FileStoragePath, logger.Log)
	memoryStorage := storage.NewMemoryStorage(keeper, logger.Log)

	controller := NewBaseController(memoryStorage, option, logger.Log, logger.RequestLogger, compressor.GzipMiddleware)

	curentFunc := controller.shortenURL
	if isJSONTest {
		curentFunc = controller.shortenJSON
	}

	handler := http.HandlerFunc(compressor.GzipMiddleware(curentFunc))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		if isJSONTest {
			require.JSONEq(t, successBody, string(b))
		} else {
			require.Equal(t, successBody, string(b))
		}
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		if isJSONTest {
			require.JSONEq(t, successBody, string(b))
		} else {
			require.Equal(t, successBody, string(b))
		}

	})
}
