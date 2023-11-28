package storage

import (
	"fmt"
	"testing"

	"github.com/wurt83ow/tinyurl/internal/config"
	"github.com/wurt83ow/tinyurl/internal/logger"
	models "github.com/wurt83ow/tinyurl/internal/models"
)

type test struct {
	keeper  *MockKeeper
	nLogger *logger.Logger
}

func beforeEach(t *testing.T) test {
	option := config.NewOptions()
	option.ParseFlags()

	nLogger, _ := logger.NewLogger(option.LogLevel())

	data := make(StorageURL)
	data["some_key"] = models.DataURL{
		UUID: "some_UUID", ShortURL: "http://localhost:8080/VajcMuGMY9h",
		OriginalURL: "https://www.google.com", UserID: "some_user_UUID"}

	users := make(StorageUser)
	users["some_key"] = models.DataUser{UUID: "some_user_UUID",
		Email: "test@gmail.com", Hash: []byte("some_hash"), Name: "some_name"}

	keeper := NewMockKeeper(t)
	keeper.On("Load").Return(data, nil)
	keeper.On("LoadUsers").Return(users, nil)

	return test{
		keeper:  keeper,
		nLogger: nLogger,
	}
}

func TestGetBaseConnection(t *testing.T) {
	test := beforeEach(t)
	test.keeper.On("Ping").Return(true)

	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	got := memStorage.GetBaseConnection()
	fmt.Println(got)
	if !got {
		t.Errorf("GetBaseConnection return %v; want true", got)
	}

	test.keeper.On("Ping").Return(false)
	memStorage = NewMemoryStorage(nil, test.nLogger)
	got = memStorage.GetBaseConnection()
	fmt.Println(got)
	if got {
		t.Errorf("GetBaseConnection return %v; want false", got)
	}
}

func TestSaveBatch(t *testing.T) {

	data := make(map[string]models.DataURL)
	test := beforeEach(t)
	test.keeper.On("SaveBatch", data).Return(nil)

	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	got := memStorage.SaveBatch(data)
	if got != nil {
		t.Errorf("SaveBatch return %v; want nil", got)
	}

	test.keeper.On("SaveBatch", data).Return(nil)
	memStorage = NewMemoryStorage(nil, test.nLogger)
	got = memStorage.SaveBatch(data)
	if got != nil {
		t.Errorf("SaveBatch return %v; want nil", got)
	}
}

func TestSaveUser(t *testing.T) {
	data := models.DataUser{}
	test := beforeEach(t)
	test.keeper.On("SaveUser", "some_key", data).Return(data, nil)

	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	_, err := memStorage.SaveUser("some_key", data)

	if err != nil {
		t.Errorf("SaveUser return error %v", err)
	}

	memStorage = NewMemoryStorage(nil, test.nLogger)
	_, err = memStorage.SaveUser("some_key", data)

	if err != nil {
		t.Errorf("SaveUser return error %v", err)
	}
}

func TestGetURL(t *testing.T) {

	test := beforeEach(t)

	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	_, err := memStorage.GetURL("some_key")

	if err != nil {
		t.Errorf("GetURL return error %v", err)
	}

	memStorage = NewMemoryStorage(test.keeper, test.nLogger)
	data, err := memStorage.GetURL("fake_key")

	if err == nil {
		t.Errorf("GetURL return value %v; want err", data)
	}
}

func TestGetUser(t *testing.T) {

	test := beforeEach(t)

	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	_, err := memStorage.GetUser("some_key")

	if err != nil {
		t.Errorf("GetUser return error %v", err)
	}

	memStorage = NewMemoryStorage(test.keeper, test.nLogger)
	data, err := memStorage.GetUser("fake_key")

	if err == nil {
		t.Errorf("GetUser return value %v; want err", data)
	}
}

func TestInsertURL(t *testing.T) {

	test := beforeEach(t)
	data := models.DataURL{
		UUID: "UUID_insertURL", ShortURL: "some_short",
		OriginalURL: "some_origin"}

	test.keeper.On("Save", "insert_key", data).Return(data, nil)
	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	data, err := memStorage.InsertURL("insert_key", data)

	if err != nil || data.UUID != "UUID_insertURL" {
		t.Errorf("InsertURL return error %v", err)
	}
}

func TestInsertUser(t *testing.T) {

	test := beforeEach(t)
	data := models.DataUser{UUID: "UUID_insertUSER",
		Email: "test@gmail.com", Hash: []byte("some_hash"), Name: "some_name"}

	test.keeper.On("SaveUser", "insert_key", data).Return(data, nil)
	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	data, err := memStorage.InsertUser("insert_key", data)

	if err != nil || data.UUID != "UUID_insertUSER" {
		t.Errorf("InsertUser return error %v", err)
	}
}

func TestInsertBatch(t *testing.T) {

	test := beforeEach(t)
	data := make(StorageURL)
	entry := models.DataURL{
		UUID: "UUID_insertURL", ShortURL: "some_short",
		OriginalURL: "some_origin"}
	data["batch_key"] = entry

	test.keeper.On("SaveBatch", data).Return(nil)
	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	err := memStorage.InsertBatch(data)

	if err != nil {
		t.Errorf("InsertBatch return error %v", err)
	}

}

func TestGetUserURLs(t *testing.T) {

	test := beforeEach(t)

	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	data := memStorage.GetUserURLs("some_user_UUID")

	if len(data) == 0 {
		t.Errorf("GetUserURLs return 0 entry; want > 0 entry")
	}

	memStorage = NewMemoryStorage(test.keeper, test.nLogger)
	data = memStorage.GetUserURLs("fake_key_user_UUID")

	if len(data) > 0 {
		t.Errorf("GetUserURLs return > 0 entry; want 0 entry")
	}
}

func TestSaveURL(t *testing.T) {
	data := models.DataURL{}
	test := beforeEach(t)
	test.keeper.On("Save", "some_key", data).Return(data, nil)

	memStorage := NewMemoryStorage(test.keeper, test.nLogger)
	_, err := memStorage.SaveURL("some_key", data)

	if err != nil {
		t.Errorf("SaveURL return error %v", err)
	}

	memStorage = NewMemoryStorage(nil, test.nLogger)
	_, err = memStorage.SaveURL("some_key", data)

	if err != nil {
		t.Errorf("SaveURL return error %v", err)
	}
}
