package controllers

import (
	"github.com/stretchr/testify/mock"
	"github.com/wurt83ow/tinyurl/internal/models"
	"github.com/wurt83ow/tinyurl/internal/storage"
)

// MockKeeper - mock implementation for the Keeper interface
type MockKeeper struct {
	mock.Mock
}

// Load - mock method for loading data
func (m *MockKeeper) Load() (storage.StorageURL, error) {
	args := m.Called()
	return args.Get(0).(storage.StorageURL), args.Error(1)
}

// LoadUsers - mock method for loading users
func (m *MockKeeper) LoadUsers() (storage.StorageUser, error) {
	args := m.Called()
	return args.Get(0).(storage.StorageUser), args.Error(1)
}

// GetUsersCount - mock method for getting the number of users and URLs
func (m *MockKeeper) GetUsersCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// GetURLsCount - mock method for getting the number of users and URLs
func (m *MockKeeper) GetURLsCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// Save - mock method for saving data
func (m *MockKeeper) Save(k string, v models.DataURL) (models.DataURL, error) {
	args := m.Called(k, v)
	return args.Get(0).(models.DataURL), args.Error(1)
}

// SaveBatch - mock method for saving a data batch
func (m *MockKeeper) SaveUser(k string, v models.DataUser) (models.DataUser, error) {
	args := m.Called(k, v)
	return args.Get(0).(models.DataUser), args.Error(1)
}

// SaveBatch - mock method for saving a data batch
func (m *MockKeeper) SaveBatch(storageURL storage.StorageURL) error {
	args := m.Called(storageURL)
	return args.Error(0)
}

// UpdateBatch - mock method for updating a data batch
func (m *MockKeeper) UpdateBatch(deleteURLs ...models.DeleteURL) error {
	args := m.Called(deleteURLs)
	return args.Error(0)
}

// Ping - mock method for checking the connection
func (m *MockKeeper) Ping() bool {
	args := m.Called()
	return args.Bool(0)
}

// Close - mock method for closing a connection
func (m *MockKeeper) Close() bool {
	args := m.Called()
	return args.Bool(0)
}
