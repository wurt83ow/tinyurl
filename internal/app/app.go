// Package app provides the main application logic for running the tinyurl server.
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

// Run starts the tinyurl server.
func Run() error {

	// Parse command line flags and environment variables for configuration options
	option := config.NewOptions()
	option.ParseFlags()

	// Initialize logger
	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		return err
	}

	// Initialize storage keeper based on configuration
	var keeper storage.Keeper = nil
	if option.DataBaseDSN() != "" {
		keeper = bdkeeper.NewBDKeeper(option.DataBaseDSN, nLogger)
	} else if option.FileStoragePath() != "" {
		keeper = filekeeper.NewFileKeeper(option.FileStoragePath, nLogger)
	}

	// Close the keeper when the function exits
	if keeper != nil {
		defer keeper.Close()
	}

	// Create a background context
	ctx := context.Background()

	// Initialize memory storage with the chosen keeper and logger
	memoryStorage := storage.NewMemoryStorage(keeper, nLogger)

	// Initialize worker, authorization, and controller
	worker := worker.NewWorker(nLogger, memoryStorage)
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)
	controller := controllers.NewBaseController(memoryStorage, option, nLogger, worker, authz)

	// Initialize request logger middleware
	reqLog := middleware.NewReqLog(nLogger)

	// Start the worker
	worker.Start(ctx)

	// Create a new Chi router
	r := chi.NewRouter()

	// Use request logger middleware and Gzip middleware
	r.Use(reqLog.RequestLogger)
	r.Use(middleware.GzipMiddleware)
  
  // Mount the controller routes
	r.Mount("/", controller.Route())

	// Get the server address from the configuration
	flagRunAddr := option.RunAddr()
	nLogger.Info("Running server", zap.String("address", flagRunAddr))

	// Allow some time for the server to start before profiling
	time.Sleep(50 * time.Millisecond)

	// Create a memory profiling log file
	memory, err := os.Create(`result.pprof`)
	if err != nil {
		panic(err)
	}
	defer memory.Close()
	// Get statistics on memory usage
	runtime.GC()

	// Write heap profile to the log file
	if err := pprof.WriteHeapProfile(memory); err != nil {
		panic(err)
	}

	// Start the HTTP server
	return http.ListenAndServe(flagRunAddr, r)

}
