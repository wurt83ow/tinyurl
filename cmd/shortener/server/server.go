package server

import (
	"net/http"

	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"

	"github.com/wurt83ow/tinyurl/internal/controllers"
)

func Run() error {
	host := "localhost"
	port := "8080"
	memoryStorage := storage.NewMemoryStorage()
	handler := controllers.NewBaseController(memoryStorage)
	return http.ListenAndServe(host+":"+port, http.HandlerFunc(handler.Webhook))
}
