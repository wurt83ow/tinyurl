package storage

import (
	"testing"
)

func TestGet(t *testing.T) {

	storage := new(MockStorage)

	// Говорим что на первый вызов Get ожидаем ключ k1 и возвращаем v1
	storage.On("Get", "k1").Return("v1", nil)

	// Говорим что на второй вызов Get ожидаем ключ k2 и возвращаем v2
	storage.On("Get", "k2").Return("v2", nil)

	got, err := storage.Get("k1")
	if got != "v1" || err != nil {
		t.Errorf("Get returns (%v, %v)", got, err)
	}
	got, err = storage.Get("k2")
	if got != "v2" || err != nil {
		t.Errorf("Get returns (%v, %v)", got, err)
	}

	// Проверяем что было вызвано всё, что устанавливали в моке
	storage.AssertExpectations(t)

}

func TestInsert(t *testing.T) {

	storage := new(MockStorage)

	storage.On("Insert", "k1", "v1").Return(nil)
	storage.On("Insert", "k2", "v2").Return(nil)

	err := storage.Insert("k1", "v1")
	if err != nil {
		t.Errorf("Insert returns error: %s", err)
	}
	err = storage.Insert("k2", "v2")
	if err != nil {
		t.Errorf("Insert returns error: %s", err)
	}

	// Проверяем что было вызвано всё, что устанавливали в моке
	storage.AssertExpectations(t)
}
