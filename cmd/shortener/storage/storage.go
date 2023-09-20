package storage

import (
	"errors"

	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrConflict указывает на конфликт данных в хранилище.
var ErrConflict = errors.New("data conflict")

type StorageURL = map[string]models.DataURL
type StorageUser = map[string]models.DataUser

type MemoryStorage struct {
	data   StorageURL
	users  StorageUser
	keeper Keeper
	log    Log
}

type Log interface {
	Info(string, ...zapcore.Field)
}

type Keeper interface {
	Load() (StorageURL, error)
	LoadUsers() (StorageUser, error)
	Save(string, models.DataURL) (models.DataURL, error)
	SaveUser(string, models.DataUser) (models.DataUser, error)
	SaveBatch(StorageURL) error
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
			log.Info("cannot decode JSON file: ", zap.Error(err))
		}
		users, err = keeper.LoadUsers()

		if err != nil {
			log.Info("cannot decode JSON file: ", zap.Error(err))
		}
	}

	return &MemoryStorage{
		data:   data,
		users:  users,
		keeper: keeper,
		log:    log,
	}
}

func (s *MemoryStorage) Insert(k string, v models.DataURL) (models.DataURL, error) {

	nv, err := s.Save(k, v)
	if err != nil {
		return nv, err
	}

	s.data[k] = nv

	return nv, nil
}

func (s *MemoryStorage) Get(k string) (models.DataURL, error) {
	v, exists := s.data[k]
	if !exists {
		return models.DataURL{}, errors.New("value with such key doesn't exist")
	}
	return v, nil
}

func (s *MemoryStorage) GetUserURLs(userID string) []models.ResponseUserURLs {
	var data []models.ResponseUserURLs

	for _, url := range s.data {
		if url.UserID == userID {
			data = append(data, models.ResponseUserURLs{OriginalURL: url.OriginalURL, ShortURL: url.ShortURL})
		}
	}

	return data
}

func (s *MemoryStorage) Save(k string, v models.DataURL) (models.DataURL, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.Save(k, v)
}

// SaveUser implements controllers.Storage.
func (s *MemoryStorage) SaveUser(k string, v models.DataUser) (models.DataUser, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.SaveUser(k, v)
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

func (s *MemoryStorage) SaveBatch(stg StorageURL) error {
	if s.keeper == nil {
		return nil
	}

	return s.keeper.SaveBatch(stg)

}

func (s *MemoryStorage) InsertUser(k string, v models.DataUser) (models.DataUser, error) {

	nv, err := s.SaveUser(k, v)
	if err != nil {
		return nv, err
	}

	s.users[k] = nv

	return nv, nil
}

func (s *MemoryStorage) GetUser(k string) (models.DataUser, error) {

	v, exists := s.users[k]

	if !exists {
		return models.DataUser{}, errors.New("value with such key doesn't exist")
	}
	return v, nil
}

func (s *MemoryStorage) GetBaseConnection() bool {
	if s.keeper == nil {
		return false
	}
	return s.keeper.Ping()
}
