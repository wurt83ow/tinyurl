package server

import (
	"controllers"
	"net/http"
	"storage"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "8080"
)

func Run() error {
	memoryStorage := storage.NewMemoryStorage()
	handler := controllers.NewHandler(memoryStorage)
	return http.ListenAndServe(CONN_HOST+":"+CONN_PORT, http.HandlerFunc(handler.Webhook))
}
