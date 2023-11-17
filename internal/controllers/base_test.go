package controllers

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authz "github.com/wurt83ow/tinyurl/internal/authorization"
	"github.com/wurt83ow/tinyurl/internal/bdkeeper"
	"github.com/wurt83ow/tinyurl/internal/config"
	"github.com/wurt83ow/tinyurl/internal/filekeeper"
	"github.com/wurt83ow/tinyurl/internal/logger"
	compressor "github.com/wurt83ow/tinyurl/internal/middleware"
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

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			// check the correctness of the received response body if we expect it
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, strings.TrimSpace(w.Body.String()), "Тело ответа не совпадает с ожидаемым")
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

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")

			assert.Equal(t, tc.location, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым")
		})
	}
}

func TestGzipShortenJSON(t *testing.T) {

	// describe the body being transmitted
	userReq := `{         
        "url": "https://practicum.yandex.ru/"
    }`

	// describe the expected response body for a successful request
	successBody := `{"result":"http://localhost:8080/nOykhckC3Od"}`

	testGzipCompression(t, userReq, successBody, true)
}

func TestGzipTestShortenURL(t *testing.T) {
	// describe the body being transmitted
	userReq := "https://practicum.yandex.ru/"

	// describe the expected response body for a successful request
	successBody := "http://localhost:8080/nOykhckC3Od"

	testGzipCompression(t, userReq, successBody, false)
}

func testGzipCompression(t *testing.T, userReq string, successBody string, isJSONTest bool) {

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

	curentFunc := controller.shortenURL
	if isJSONTest {
		curentFunc = controller.shortenJSON
	}

	handler := compressor.GzipMiddleware(http.HandlerFunc(curentFunc))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(userReq))
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
		buf := bytes.NewBufferString(userReq)
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
