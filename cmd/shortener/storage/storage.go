package storage

import "errors"

type Storage interface {
	Insert(k string, v string) error
	Get(k string) (string, error)
}

type MemoryStorage struct {
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (s *MemoryStorage) Insert(k string, v string) error {
	s.data[k] = v

	return nil
}

func (s *MemoryStorage) Get(k string) (string, error) {
	v, exists := s.data[k]
	if !exists {
		return "", errors.New("value with such key doesn't exist.")
	}
	return v, nil
}
