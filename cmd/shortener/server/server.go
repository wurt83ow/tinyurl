package server

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi"

	"github.com/wurt83ow/tinyurl/cmd/shortener/config"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"
	"github.com/wurt83ow/tinyurl/internal/bdkeeper"
	"github.com/wurt83ow/tinyurl/internal/controllers"
	"github.com/wurt83ow/tinyurl/internal/filekeeper"
	"github.com/wurt83ow/tinyurl/internal/logger"
	"github.com/wurt83ow/tinyurl/internal/middleware"
)

func Run() error {

	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		return err
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

	controller := controllers.NewBaseController(memoryStorage, option, nLogger)
	reqLog := middleware.NewReqLog(nLogger)

	r := chi.NewRouter()
	r.Use(reqLog.RequestLogger)
	r.Use(middleware.GzipMiddleware)
	r.Mount("/", controller.Route())

	flagRunAddr := option.RunAddr()
	nLogger.Info("Running server", zap.String("address", flagRunAddr))

	return http.ListenAndServe(flagRunAddr, r)
}
