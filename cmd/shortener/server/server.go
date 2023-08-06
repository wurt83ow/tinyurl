package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"

	"github.com/wurt83ow/tinyurl/internal/controllers"
)

func Run() error {
	host := "localhost"
	port := "8080"
	memoryStorage := storage.NewMemoryStorage()
	controller := controllers.NewBaseController(memoryStorage)

	r := chi.NewRouter()
	r.Mount("/", controller.Route())
	return http.ListenAndServe(host+":"+port, r)
}
