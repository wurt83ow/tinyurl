package server

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/wurt83ow/tinyurl/cmd/shortener/config"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"
	"github.com/wurt83ow/tinyurl/internal/compressor"
	"github.com/wurt83ow/tinyurl/internal/controllers"
	"github.com/wurt83ow/tinyurl/internal/logger"
)

func Run() error {

	option := config.NewOptions()
	option.ParseFlags()

	if err := logger.Initialize(option.LogLevel()); err != nil {
		return err
	}

	memoryStorage := storage.NewMemoryStorage()
	controller := controllers.NewBaseController(memoryStorage, option, logger.Log, logger.RequestLogger, compressor.GzipMiddleware)

	r := chi.NewRouter()
	r.Mount("/", controller.Route())

	flagRunAddr := option.RunAddr()
	logger.Log.Info("Running server", zap.String("address", flagRunAddr))

	return http.ListenAndServe(flagRunAddr, r)
}
