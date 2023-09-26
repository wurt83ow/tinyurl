package storage

import (
	"testing"
)

func TestGet(t *testing.T) {

	storage := new(MockStorage)

	storage.On("Get", "k1").Return("v1", nil)
	storage.On("Get", "k2").Return("v2", nil)

	got, err := storage.Get("k1")
	if got != "v1" || err != nil {
		t.Errorf("Get returns (%v, %v)", got, err)
	}
	got, err = storage.Get("k2")
	if got != "v2" || err != nil {
		t.Errorf("Get returns (%v, %v)", got, err)
	}

	// check that everything that was installed in the mock was called
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

	// check that everything that was installed in the mock was called
	storage.AssertExpectations(t)
}
