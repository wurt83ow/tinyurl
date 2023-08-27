package storage

import (
	"encoding/json"
	"errors"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type DataURL struct {
	UUID        int64  `json:"result"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type MemoryStorage struct {
	data map[string]string
	path func() string
	log  Log
}

type Log interface {
	Info(string, ...zapcore.Field)
}

func NewMemoryStorage(path func() string, log Log) *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
		path: path,
		log:  log,
	}
}

func (s *MemoryStorage) Insert(k string, v string) error {
	s.data[k] = v
	err := s.save()
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

func (s *MemoryStorage) Load() error {
	dataFile := s.path()
	if _, err := os.Stat(dataFile); err != nil {
		s.log.Info("file not found", zap.Error(err))
		return nil
	}

	loadFrom, err := os.Open(dataFile)

	if err != nil {
		s.log.Info("Empty key/value store!", zap.Error(err))
		return err
	}
	defer loadFrom.Close()

	decoder := json.NewDecoder(loadFrom)
	for decoder.More() {
		var m DataURL
		err := decoder.Decode(&m)
		s.data[m.ShortURL] = m.OriginalURL

		if err != nil {
			s.log.Info("cannot decode JSON file", zap.Error(err))
		}
	}

	return nil
}

func (s *MemoryStorage) save() error {

	dataFile := s.path()

	if _, err := os.Stat(dataFile); err == nil {
		err := os.Remove(dataFile)
		if err != nil {
			s.log.Info("Cannot remove file", zap.Error(err))
		}
	}

	saveTo, err := os.Create(dataFile)
	if err != nil {
		s.log.Info("Cannot create file", zap.Error(err))
		return err
	}
	defer saveTo.Close()

	var i int64 = 0

	for k, v := range s.data {
		i++
		data := DataURL{
			UUID: i, ShortURL: k,
			OriginalURL: v}
		encoder := json.NewEncoder(saveTo)
		err = encoder.Encode(data)
		if err != nil {
			s.log.Info("cannot encode JSON data", zap.Error(err))
			return err
		}
	}

	return nil
}
