package worker

import (
	"context"
	"sync"
	"time"

	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log interface {
	Warn(msg string, fields ...zapcore.Field)
	Info(string, ...zapcore.Field)
}

type Storage interface {
	DeleteURLs(delUrls ...models.DeleteURL) error
}

type worker struct {
	wg         *sync.WaitGroup
	cancelFunc context.CancelFunc
	log        Log
	storage    Storage
	jobChan    chan models.DeleteURL
	result     []models.DeleteURL //chan interface{}
}

// type workType string

type Worker interface {
	Start(pctx context.Context)
	Stop()
	// QueueTask(task string, workDuration time.Duration) error
}

func NewWorker(log Log, storage Storage) Worker {
	w := worker{
		wg:      new(sync.WaitGroup),
		log:     log,
		storage: storage,
		jobChan: make(chan models.DeleteURL, 1024), // set the channel buffer to 1024 messages
		result:  make([]models.DeleteURL, 0),       //make(chan interface{}, 1024)
	}

	return &w
}

func (w *worker) Start(pctx context.Context) {
	w.log.Warn("Start worker")
	ctx, canselFunc := context.WithCancel(pctx)
	w.cancelFunc = canselFunc
	w.wg.Add(1)
	go w.spawnWorkers(ctx)
}

func (w *worker) Stop() {
	w.log.Warn("Stop worker")
	w.cancelFunc()
	w.wg.Wait()
	w.log.Warn("All workers exited!")
}

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

func (w *worker) doWork(ctx context.Context) {
	w.log.Warn("I'm here")

	if len(w.result) != 0 {

		// save all incoming messages at once
		err := w.storage.DeleteURLs(w.result...)
		if err != nil {
			w.log.Info("cannot save delUrls", zap.Error(err))
			// not delete messages, we'll try to send them a little later

		}
		// erase successfully sent messages
		w.result = nil
	}

	// rnd := rand.Int63()
	// w.storage.Set(strconv.FormatInt(rnd, 36), strconv.FormatInt(rnd, 10))
}
