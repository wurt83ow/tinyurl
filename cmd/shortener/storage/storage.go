package storage

import (
	"errors"

	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type StorageURL = map[string]models.DataURL

type MemoryStorage struct {
	data   StorageURL
	keeper Keeper
	log    Log
}

type Log interface {
	Info(string, ...zapcore.Field)
}

type Keeper interface {
	Load() (StorageURL, error)
	Save(StorageURL) error
	Ping() bool
	Close() bool
}

func NewMemoryStorage(keeper Keeper, log Log) *MemoryStorage {

	data := make(StorageURL)

	if keeper != nil {
		var err error
		data, err = keeper.Load()
		if err != nil {
			log.Info("cannot decode JSON file", zap.Error(err))
		}
	}

	return &MemoryStorage{
		data:   data,
		keeper: keeper,
		log:    log,
	}
}

func (s *MemoryStorage) Insert(k string, v models.DataURL, save bool) error {
	s.data[k] = v

	if save {
		s.Save()
	}

	return nil
}

func (s *MemoryStorage) Get(k string) (models.DataURL, error) {
	v, exists := s.data[k]
	if !exists {
		return models.DataURL{}, errors.New("value with such key doesn't exist")
	}
	return v, nil
}

func (s *MemoryStorage) Save() bool {
	if s.keeper == nil {
		return true
	}

	err := s.keeper.Save(s.data)

	if err != nil {
		s.log.Info("cannot insert value to JSON file", zap.Error(err))
		return false
	}

	return true
}

func (s *MemoryStorage) GetBaseConnection() bool {
	if s.keeper == nil {
		return false
	}
	return s.keeper.Ping()
}
