// Package storage provides an in-memory storage implementation with CRUD operations for URL and user data.
// It includes interfaces and a MemoryStorage type implementing these interfaces.
package storage

import (
	"errors"
	"strings"
	"sync"

	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrConflict indicates a data conflict in the store.
var ErrConflict = errors.New("data conflict")

// StorageURL represents a mapping of string keys to DataURL values.
type StorageURL = map[string]models.DataURL

// StorageUser represents a mapping of string keys to DataUser values.
type StorageUser = map[string]models.DataUser

// Log is an interface representing a logger with Info method.
type Log interface {
	Info(string, ...zapcore.Field)
}

// MemoryStorage is an in-memory storage implementation with CRUD operations for URL and user data.
type MemoryStorage struct {
	data   StorageURL
	users  StorageUser
	keeper Keeper
	log    Log
	dmx    sync.RWMutex
	umx    sync.RWMutex
}

// Keeper is an interface representing methods for loading, saving, and updating data in storage.
type Keeper interface {
	Load() (StorageURL, error)
	LoadUsers() (StorageUser, error)
	GetUsersCount() (int, error)
	GetURLsCount() (int, error)
	Save(string, models.DataURL) (models.DataURL, error)
	SaveUser(string, models.DataUser) (models.DataUser, error)
	SaveBatch(StorageURL) error
	UpdateBatch(...models.DeleteURL) error
	Ping() bool
	Close() bool
}

// NewMemoryStorage creates a new MemoryStorage instance with the provided Keeper and logger.
func NewMemoryStorage(keeper Keeper, log Log) *MemoryStorage {
	data := make(StorageURL)
	users := make(StorageUser)

	if keeper != nil {
		var err error
		data, err = keeper.Load()
		if err != nil {
			log.Info("cannot load url data: ", zap.Error(err))
		}

		users, err = keeper.LoadUsers()
		if err != nil {
			log.Info("cannot load user data: ", zap.Error(err))
		}
	}

	return &MemoryStorage{
		data:   data,
		users:  users,
		keeper: keeper,
		log:    log,
	}
}

// GetUsersCount Gets the number of users from Keeper.
func (s *MemoryStorage) GetUsersCount() (int, error) {
	return s.keeper.GetUsersCount()
}

// GetURLsCount Gets the number of URLs from Keeper.
func (s *MemoryStorage) GetURLsCount() (int, error) {
	return s.keeper.GetURLsCount()
}

// InsertURL inserts a new DataURL into the storage with the specified key.
func (s *MemoryStorage) InsertURL(k string, v models.DataURL) (models.DataURL, error) {
	nv, err := s.SaveURL(k, v)
	if err != nil {
		return nv, err
	}

	s.dmx.Lock()
	defer s.dmx.Unlock()

	s.data[k] = nv

	return nv, nil
}

// InsertUser inserts a new DataUser into the storage with the specified key.
func (s *MemoryStorage) InsertUser(k string, v models.DataUser) (models.DataUser, error) {
	nv, err := s.SaveUser(k, v)
	if err != nil {
		return nv, err
	}

	s.umx.Lock()
	defer s.umx.Unlock()

	s.users[k] = nv

	return nv, nil
}

// InsertBatch inserts a batch of DataURL values into the storage.
func (s *MemoryStorage) InsertBatch(stg StorageURL) error {
	for k, v := range stg {
		s.data[k] = v
	}

	err := s.SaveBatch(stg)
	if err != nil {
		return err
	}

	return nil
}

// GetURL retrieves a DataURL from the storage with the specified key.
func (s *MemoryStorage) GetURL(k string) (models.DataURL, error) {
	s.dmx.RLock()
	defer s.dmx.RUnlock()

	v, exists := s.data[k]

	if !exists {
		return models.DataURL{}, errors.New("value with such key doesn't exist")
	}

	return v, nil
}

// GetUser retrieves a DataUser from the storage with the specified key.
func (s *MemoryStorage) GetUser(k string) (models.DataUser, error) {
	s.umx.RLock()
	defer s.umx.RUnlock()

	v, exists := s.users[k]
	if !exists {
		return models.DataUser{}, errors.New("value with such key doesn't exist")
	}

	return v, nil
}

// GetUserURLs retrieves a slice of DataURLite for a specific user from the storage.
func (s *MemoryStorage) GetUserURLs(userID string) []models.DataURLite {
	var data []models.DataURLite

	s.dmx.RLock()
	defer s.dmx.RUnlock()
	for _, u := range s.data {
		if u.UserID == userID {
			data = append(data, models.DataURLite{
				OriginalURL: u.OriginalURL, ShortURL: u.ShortURL})
		}
	}

	return data
}

// SaveURL saves a DataURL to the storage using the provided key.
func (s *MemoryStorage) SaveURL(k string, v models.DataURL) (models.DataURL, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.Save(k, v)
}

// DeleteURLs deletes URLs from the storage based on the provided delete URLs.
func (s *MemoryStorage) DeleteURLs(delUrls ...models.DeleteURL) error {
	if s.keeper == nil {
		return nil
	}

	err := s.keeper.UpdateBatch(delUrls...)
	if err != nil {
		return err
	}

	s.dmx.RLock()
	defer s.dmx.RUnlock()

	for _, u := range delUrls {
		for _, k := range u.ShortURLs {
			cs := s.data[k]

			if cs.UserID == u.UserID && strings.Contains(cs.ShortURL, k) {
				s.data[k] = models.DataURL{UUID: cs.UUID, ShortURL: cs.ShortURL,
					OriginalURL: cs.OriginalURL, UserID: cs.UserID, DeletedFlag: true}
			}
		}
	}

	return nil
}

// SaveUser saves a DataUser to the storage using the provided key.
func (s *MemoryStorage) SaveUser(k string, v models.DataUser) (models.DataUser, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.SaveUser(k, v)
}

// SaveBatch saves a batch of DataURL values to the storage.
func (s *MemoryStorage) SaveBatch(stg StorageURL) error {
	if s.keeper == nil {
		return nil
	}

	return s.keeper.SaveBatch(stg)
}

// GetBaseConnection checks the connectivity of the underlying storage keeper.
func (s *MemoryStorage) GetBaseConnection() bool {
	if s.keeper == nil {
		return false
	}

	return s.keeper.Ping()
}
