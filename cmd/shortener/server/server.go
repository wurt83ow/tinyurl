package server

import (
	"net/http"

	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"

	"github.com/wurt83ow/tinyurl/internal/controllers"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "8080"
)

func Run() error {
	memoryStorage := storage.NewMemoryStorage()
	handler := controllers.NewBaseController(memoryStorage)
	return http.ListenAndServe(CONN_HOST+":"+CONN_PORT, http.HandlerFunc(handler.Webhook))
}
