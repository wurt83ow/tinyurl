package app

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/go-chi/chi"
	authz "github.com/wurt83ow/tinyurl/internal/authorization"
	"github.com/wurt83ow/tinyurl/internal/bdkeeper"
	"github.com/wurt83ow/tinyurl/internal/config"
	"github.com/wurt83ow/tinyurl/internal/controllers"
	"github.com/wurt83ow/tinyurl/internal/filekeeper"
	"github.com/wurt83ow/tinyurl/internal/logger"
	"github.com/wurt83ow/tinyurl/internal/middleware"
	"github.com/wurt83ow/tinyurl/internal/storage"
	"github.com/wurt83ow/tinyurl/internal/worker"
	"go.uber.org/zap"
)

// func foo() {
// 	maxSize := 50000000
// 	// полезная нагрузка
// 	for i := 0; i < 10; i++ {
// 		s := make([]byte, maxSize)
// 		if s == nil {
// 			fmt.Println("Operation failed!")
// 		}
// 		time.Sleep(50 * time.Millisecond)
// 	}
// }

func Run() error {
	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		return err
	}

	var keeper storage.Keeper = nil
	if option.DataBaseDSN() != "" {
		keeper = bdkeeper.NewBDKeeper(option.DataBaseDSN, nLogger)
	} else if option.FileStoragePath() != "" {
		keeper = filekeeper.NewFileKeeper(option.FileStoragePath, nLogger)
	}

	if keeper != nil {
		defer keeper.Close()
	}

	ctx := context.Background()

	memoryStorage := storage.NewMemoryStorage(keeper, nLogger)

	worker := worker.NewWorker(nLogger, memoryStorage)
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)
	controller := controllers.NewBaseController(memoryStorage, option, nLogger, worker, authz)

	reqLog := middleware.NewReqLog(nLogger)

	worker.Start(ctx)
	r := chi.NewRouter()

	r.Use(reqLog.RequestLogger)
	r.Use(middleware.GzipMiddleware)

	// r.Mount("/debug", middleware.Profiler())
	r.Mount("/", controller.Route())

	flagRunAddr := option.RunAddr()
	nLogger.Info("Running server", zap.String("address", flagRunAddr))

	time.Sleep(50 * time.Millisecond)
	// создаём файл журнала профилирования памяти
	memory, err := os.Create(`result.pprof`)
	if err != nil {
		panic(err)
	}
	defer memory.Close()
	runtime.GC() // получаем статистику по использованию памяти

	// go foo()
	if err := pprof.WriteHeapProfile(memory); err != nil {
		panic(err)
	}

	return http.ListenAndServe(flagRunAddr, r)
}
