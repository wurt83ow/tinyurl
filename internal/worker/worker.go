// Package worker provides a background worker for deleting URLs from storage.
// It includes interfaces and an implementation for managing jobs.
package worker

import (
	"context"
	"sync"
	"time"

	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is an interface representing a logger with Warn and Info methods.
type Log interface {
	Warn(msg string, fields ...zapcore.Field)
	Info(string, ...zapcore.Field)
}

// Storage is an interface representing a data storage with a method to delete URLs.
type Storage interface {
	DeleteURLs(delUrls ...models.DeleteURL) error
}

// Worker is an interface representing a background worker for deleting URLs.
type Worker interface {
	// Start starts the worker with the given parent context.
	Start(pctx context.Context)
	// Stop stops the worker.
	Stop()
	// Add adds a job to the worker's job channel.
	Add(models.DeleteURL)
}

// worker is an implementation of the Worker interface.
type worker struct {
	wg         *sync.WaitGroup
	cancelFunc context.CancelFunc
	log        Log
	storage    Storage
	jobChan    chan models.DeleteURL
	result     []models.DeleteURL
}

// NewWorker creates a new Worker instance with the provided logger and storage.
func NewWorker(log Log, storage Storage) Worker {
	w := worker{
		wg:      new(sync.WaitGroup),
		log:     log,
		storage: storage,
		jobChan: make(chan models.DeleteURL, 1024),
		result:  make([]models.DeleteURL, 0),
	}

	return &w
}

// Start starts the worker with the given parent context.
func (w *worker) Start(pctx context.Context) {
	w.log.Warn("Start worker")
	ctx, cancelFunc := context.WithCancel(pctx)
	w.cancelFunc = cancelFunc
	w.wg.Add(1)
	go w.spawnWorkers(ctx)
}

// Stop stops the worker.
func (w *worker) Stop() {
	w.log.Warn("Stop worker")
	w.cancelFunc()
	w.wg.Wait()
	w.log.Warn("All workers exited!")
}

// Add adds a job to the worker's job channel.
func (w *worker) Add(d models.DeleteURL) {
	w.jobChan <- d
}

// spawnWorkers is a goroutine that handles jobs and periodically performs the actual deletion.
func (w *worker) spawnWorkers(ctx context.Context) {
	defer w.wg.Done()
	w.log.Warn(" start ")
	t := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case job := <-w.jobChan:
			w.result = append(w.result, job)
		case <-t.C:
			w.doWork(ctx)
		}
	}
}

// doWork performs the actual deletion of URLs from storage.
func (w *worker) doWork(ctx context.Context) {
	if len(w.result) != 0 {
		err := w.storage.DeleteURLs(w.result...)
		if err != nil {
			w.log.Info("cannot save delUrls", zap.Error(err))
		}
		w.result = nil
	}
}
