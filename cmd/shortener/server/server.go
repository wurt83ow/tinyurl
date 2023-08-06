package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/wurt83ow/tinyurl/cmd/shortener/config"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"

	"github.com/wurt83ow/tinyurl/internal/controllers"
)

func Run() error {

	memoryStorage := storage.NewMemoryStorage()
	controller := controllers.NewBaseController(memoryStorage)
	config.ParseFlags()

	r := chi.NewRouter()
	r.Mount("/", controller.Route())

	return http.ListenAndServe(config.RunAddr(), r)
}
