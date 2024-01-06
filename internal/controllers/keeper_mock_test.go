package controllers

import (
	"github.com/stretchr/testify/mock"
	"github.com/wurt83ow/tinyurl/internal/models"
	"github.com/wurt83ow/tinyurl/internal/storage"
)

// MockKeeper - мок-реализация для интерфейса Keeper
type MockKeeper struct {
	mock.Mock
}

// Load - мок-метод для загрузки данных
func (m *MockKeeper) Load() (storage.StorageURL, error) {
	args := m.Called()
	return args.Get(0).(storage.StorageURL), args.Error(1)
}

// LoadUsers - мок-метод для загрузки пользователей
func (m *MockKeeper) LoadUsers() (storage.StorageUser, error) {
	args := m.Called()
	return args.Get(0).(storage.StorageUser), args.Error(1)
}

// GetUsersAndURLsCount - мок-метод для получения количества пользователей и URL
func (m *MockKeeper) GetUsersAndURLsCount() (int, int, error) {
	args := m.Called()
	return args.Int(0), args.Int(1), args.Error(2)
}

// Save - мок-метод для сохранения данных
func (m *MockKeeper) Save(k string, v models.DataURL) (models.DataURL, error) {
	args := m.Called(k, v)
	return args.Get(0).(models.DataURL), args.Error(1)
}

// SaveUser - мок-метод для сохранения пользователя
func (m *MockKeeper) SaveUser(k string, v models.DataUser) (models.DataUser, error) {
	args := m.Called(k, v)
	return args.Get(0).(models.DataUser), args.Error(1)
}

// SaveBatch - мок-метод для сохранения пакета данных
func (m *MockKeeper) SaveBatch(storageURL storage.StorageURL) error {
	args := m.Called(storageURL)
	return args.Error(0)
}

// UpdateBatch - мок-метод для обновления пакета данных
func (m *MockKeeper) UpdateBatch(deleteURLs ...models.DeleteURL) error {
	args := m.Called(deleteURLs)
	return args.Error(0)
}

// Ping - мок-метод для проверки соединения
func (m *MockKeeper) Ping() bool {
	args := m.Called()
	return args.Bool(0)
}

// Close - мок-метод для закрытия соединения
func (m *MockKeeper) Close() bool {
	args := m.Called()
	return args.Bool(0)
}
