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
	Save(string, models.DataURL) (models.DataURL, error)
	SaveBatch(StorageURL) error
	Ping() bool
	Close() bool
}

func NewMemoryStorage(keeper Keeper, log Log) *MemoryStorage {

	data := make(StorageURL)

	if keeper != nil {
		var err error
		data, err = keeper.Load()
		if err != nil {
			log.Info("cannot decode JSON file: ", zap.Error(err))
		}
	}

	return &MemoryStorage{
		data:   data,
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

func (s *MemoryStorage) Save(k string, v models.DataURL) (models.DataURL, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.Save(k, v)
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

func (s *MemoryStorage) GetBaseConnection() bool {
	if s.keeper == nil {
		return false
	}
	return s.keeper.Ping()
}
