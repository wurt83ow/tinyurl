// Package app provides the main application logic for running the tinyurl server.
package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	authz "github.com/wurt83ow/tinyurl/internal/authorization"
	"github.com/wurt83ow/tinyurl/internal/bdkeeper"
	"github.com/wurt83ow/tinyurl/internal/config"
	"github.com/wurt83ow/tinyurl/internal/controllers"
	pb "github.com/wurt83ow/tinyurl/internal/controllers/proto"
	"github.com/wurt83ow/tinyurl/internal/filekeeper"
	"github.com/wurt83ow/tinyurl/internal/logger"
	"github.com/wurt83ow/tinyurl/internal/middleware"
	"github.com/wurt83ow/tinyurl/internal/storage"
	"github.com/wurt83ow/tinyurl/internal/worker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	// Создайте экземпляр gRPC-сервера
	grpcServer := grpc.NewServer()

	// Регистрируйте ваш сервис gRPC
	pb.RegisterUsersServer(grpcServer, controllers.NewUsersServer())

	// Добавьте поддержку reflection API
	reflection.Register(grpcServer)
	// Создайте слушателя для gRPC
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051))
	if err != nil {
		return fmt.Errorf("failed to listen for gRPC: %v", err)
	}
	defer grpcListener.Close()

	// Ваша логика для запуска gRPC-сервера
	go func() {
		log.Printf("gRPC server is listening on port 50051")
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

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

	// Create a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Создайте HTTP-сервер
	server := &http.Server{
		Addr:    flagRunAddr,
		Handler: r,
	}

	// Используйте отдельную горутину для прослушивания сигналов ОС и грациозного завершения сервера
	go func() {
		sig := <-stop
		nLogger.Info("Received signal. Shutting down...", zap.String("signal", sig.String()))

		// Прекратить принятие новых запросов и дождитесь завершения оставшихся запросов
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Start the worker
		worker.Stop()

		// Shutdown грациозно закрывает сервер, включая ожидание завершения запросов
		if err := server.Shutdown(ctx); err != nil {
			nLogger.Info("Error shutting down server", zap.Error(err))
		}
	}()

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

	// Start the HTTP/HTTPS server
	if option.EnableHTTPS() {
		nLogger.Info("HTTPS enabled")
		return server.ListenAndServeTLS("server.crt", "server.key")
	} else {
		nLogger.Info("HTTPS disabled")
		return server.ListenAndServe()
	}

}
