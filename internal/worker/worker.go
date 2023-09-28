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
	resultChan chan models.DeleteURL
}

// type workType string

type Worker interface {
	Start(pctx context.Context)
	Stop()
	Add(models.DeleteURL)
	// QueueTask(task string, workDuration time.Duration) error
}

func NewWorker(log Log, storage Storage) Worker {
	w := worker{
		wg:         new(sync.WaitGroup),
		log:        log,
		storage:    storage,
		jobChan:    make(chan models.DeleteURL, 1024), // set the channel buffer to 1024 messages
		resultChan: make(chan models.DeleteURL, 1024),
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

func (w *worker) Add(d models.DeleteURL) {
	w.jobChan <- d
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
			w.resultChan <- job
		case <-t.C:
			w.doWork(ctx)

		}
	}
}

func (w *worker) doWork(ctx context.Context) {
	w.log.Warn("I'm here")
	count := len(w.resultChan)
	if count != 0 {
		result := make([]models.DeleteURL, count)
		for res := range w.resultChan {
			result = append(result, res)
		}
		// save all incoming messages at once
		err := w.storage.DeleteURLs(result...)
		if err != nil {
			w.log.Info("cannot save delUrls", zap.Error(err))
			// not delete messages, we'll try to send them a little later

		}
		// erase successfully sent messages
		w.resultChan = nil
	}

	// rnd := rand.Int63()
	// w.storage.Set(strconv.FormatInt(rnd, 36), strconv.FormatInt(rnd, 10))
}
