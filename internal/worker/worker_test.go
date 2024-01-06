package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap/zapcore"
)

// MockLog is a mock implementation of the Log interface for testing.
type MockLog struct {
	mock.Mock
}

func (m *MockLog) Warn(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLog) Info(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

// MockStorage is a mock implementation of the Storage interface for testing.
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) DeleteURLs(delUrls ...models.DeleteURL) error {
	args := m.Called(delUrls)
	return args.Error(0)
}

func TestWorker_StartStop(t *testing.T) {
	// Create a new instance of your worker with a mock logger and storage
	log := new(MockLog)
	storage := new(MockStorage)
	worker := NewWorker(log, storage)

	// Set up expectations for the Warn method during Start
	log.On("Warn", "Start worker", []zapcore.Field(nil)).Once()

	// Call the Start method to log "Start worker"
	worker.Start(context.Background())

	// Wait for a short duration to ensure that the worker has started
	time.Sleep(100 * time.Millisecond)

	// Assert that the Warn method was called once during Start
	log.AssertExpectations(t)

	// Set up expectations for the Warn method during Stop
	log.On("Warn", "All workers exited!", []zapcore.Field(nil)).Once()

	// Call the Stop method to log "Stop worker"
	worker.Stop()

	// Assert that all expectations were met
	log.AssertExpectations(t)
}

func TestWorker_Add(t *testing.T) {
	// Создаем новый экземпляр вашего воркера с моковым логгером и хранилищем
	log := new(MockLog)
	storage := new(MockStorage)
	worker := NewWorker(log, storage)

	// Устанавливаем ожидания для метода Warn
	log.On("Warn", mock.Anything, mock.Anything).Once()

	// Запускаем воркер
	worker.Start(context.TODO())

	// Добавляем задачу в воркер
	storage.On("AddURL", mock.Anything).Return(nil).Once()

	// Ваш остальной код теста здесь
}

func TestWorker_DoWork(t *testing.T) {
	// Создаем новый экземпляр вашего воркера с моковым логгером и хранилищем
	log := new(MockLog)
	storage := new(MockStorage)
	worker := NewWorker(log, storage)

	// Устанавливаем ожидания для метода Warn
	log.On("Warn", mock.Anything, mock.Anything).Once()

	// Запускаем воркер
	worker.Start(context.TODO())

	// Добавляем задачу в воркер
	storage.On("AddURL", mock.Anything).Return(nil).Once()

	// Ваш код, вызывающий DoWork() здесь

	// Проверяем, что метод Warn был вызван с ожидаемыми аргументами
	log.AssertExpectations(t)
}
