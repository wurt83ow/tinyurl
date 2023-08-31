package storage

import (
	"errors"

	"github.com/wurt83ow/tinyurl/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MemoryStorage struct {
	data   map[string]string
	keeper Keeper
	log    Log
}

type Log interface {
	Info(string, ...zapcore.Field)
}

type Keeper interface {
	Load() error
	Save() error
}

func NewMemoryStorage(keeper Keeper, log Log) *MemoryStorage {
	err := keeper.Load()
	if err != nil {
		logger.Log.Info("cannot decode JSON file", zap.Error(err))
	}

	return &MemoryStorage{
		data:   make(map[string]string),
		keeper: keeper,
		log:    log,
	}
}

func (s *MemoryStorage) Insert(k string, v string) error {
	s.data[k] = v
	err := s.keeper.Save()
	if err != nil {
		s.log.Info("cannot insert value to JSON file", zap.Error(err))
	}
	return nil
}

func (s *MemoryStorage) Get(k string) (string, error) {
	v, exists := s.data[k]
	if !exists {
		return "", errors.New("value with such key doesn't exist")
	}
	return v, nil
}
