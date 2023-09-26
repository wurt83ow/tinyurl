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

type StorageURL = map[string]models.DataURL
type StorageUser = map[string]models.DataUser

type Log interface {
	Info(string, ...zapcore.Field)
}

type MemoryStorage struct {
	dmx    sync.RWMutex
	umx    sync.RWMutex
	data   StorageURL
	users  StorageUser
	keeper Keeper
	log    Log
}

type Keeper interface {
	Load() (StorageURL, error)
	LoadUsers() (StorageUser, error)
	Save(string, models.DataURL) (models.DataURL, error)
	SaveUser(string, models.DataUser) (models.DataUser, error)
	SaveBatch(StorageURL) error
	UpdateBatch(...models.DeleteURL) error
	Ping() bool
	Close() bool
}

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

func (s *MemoryStorage) InsertURL(k string,
	v models.DataURL) (models.DataURL, error) {

	nv, err := s.SaveURL(k, v)
	if err != nil {
		return nv, err
	}

	s.dmx.Lock()
	defer s.dmx.Unlock()

	s.data[k] = nv

	return nv, nil
}

func (s *MemoryStorage) InsertUser(k string,
	v models.DataUser) (models.DataUser, error) {

	nv, err := s.SaveUser(k, v)
	if err != nil {
		return nv, err
	}

	s.umx.Lock()
	defer s.umx.Unlock()

	s.users[k] = nv

	return nv, nil
}

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

func (s *MemoryStorage) GetURL(k string) (models.DataURL, error) {
	s.dmx.RLock()
	defer s.dmx.RUnlock()

	v, exists := s.data[k]

	if !exists {
		return models.DataURL{}, errors.New("value with such key doesn't exist")
	}

	return v, nil
}

func (s *MemoryStorage) GetUser(k string) (models.DataUser, error) {
	s.umx.RLock()
	defer s.umx.RUnlock()

	v, exists := s.users[k]
	if !exists {
		return models.DataUser{}, errors.New("value with such key doesn't exist")
	}

	return v, nil
}

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

func (s *MemoryStorage) SaveURL(k string, v models.DataURL) (models.DataURL, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.Save(k, v)
}

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

func (s *MemoryStorage) SaveUser(k string, v models.DataUser) (models.DataUser, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.SaveUser(k, v)
}

func (s *MemoryStorage) SaveBatch(stg StorageURL) error {
	if s.keeper == nil {
		return nil
	}

	return s.keeper.SaveBatch(stg)

}

func (s *MemoryStorage) GetBaseConnection() bool {
	if s.keeper == nil {
		return false
	}

	return s.keeper.Ping()
}
