package server

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wurt83ow/tinyurl/cmd/shortener/config"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"
	"github.com/wurt83ow/tinyurl/internal/bdkeeper"
	"github.com/wurt83ow/tinyurl/internal/controllers"
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

	// fileKeeper := filekeeper.NewFileKeeper(option.FileStoragePath, nLogger)

	// go run main.go -a ":8010" -l "info" -d "host=localhost user=videos password=777 dbname=videos sslmode=disable"
	pool, err := pgxpool.New(context.Background(), option.DataBaseDSN())
	if err != nil {
		nLogger.Info("Unable to connection to database: %v", zap.Error(err))
	}
	defer pool.Close()
	nLogger.Info("Connected!")

	bdKeeper := bdkeeper.NewBDKeeper(pool, nLogger)
	err = bdKeeper.CreateTable()
	if err != nil {
		nLogger.Info("77777777777777777777777: %v", zap.Error(err))
	}

	// data, err := bdKeeper.Load()
	// if err != nil {
	// 	nLogger.Info("Unable to connection to database: %v", zap.Error(err))
	// }
	// fmt.Println(data)

	memoryStorage := storage.NewMemoryStorage(bdKeeper, nLogger)
	// memoryStorage := storage.NewMemoryStorage(fileKeeper, nLogger)

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
