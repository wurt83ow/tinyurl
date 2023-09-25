package filekeeper

import (
	"encoding/json"
	"os"

	"github.com/google/uuid"

	"github.com/wurt83ow/tinyurl/internal/models"
	"github.com/wurt83ow/tinyurl/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log interface {
	Info(string, ...zapcore.Field)
}

type FileKeeper struct {
	path func() string
	log  Log
}

// UpdateBatch implements storage.Keeper.
func (*FileKeeper) UpdateBatch(...models.DeleteURL) error {
	panic("unimplemented")
}

func NewFileKeeper(path func() string, log Log) *FileKeeper {

	addr := path()
	if addr == "" {
		log.Info("file json path is empty")
		return nil
	}

	return &FileKeeper{
		path: path,
		log:  log,
	}
}

func (kp *FileKeeper) Load() (storage.StorageURL, error) {
	dataFile := kp.path()
	data := make(storage.StorageURL)

	if _, err := os.Stat(dataFile); err != nil {
		kp.log.Info("file not found: ", zap.Error(err))
		return data, err
	}

	loadFrom, err := os.Open(dataFile)

	if err != nil {
		kp.log.Info("Empty key/value store!: ", zap.Error(err))
		return data, err
	}
	defer loadFrom.Close()

	decoder := json.NewDecoder(loadFrom)
	for decoder.More() {
		var m models.DataURL
		err := decoder.Decode(&m)
		data[m.ShortURL] = m

		if err != nil {
			kp.log.Info("cannot decode JSON file: ", zap.Error(err))
		}
	}

	return data, nil
}

func (kp *FileKeeper) LoadUsers() (storage.StorageUser, error) {
	dataFile := kp.path()
	data := make(storage.StorageUser)

	if _, err := os.Stat(dataFile); err != nil {
		kp.log.Info("file not found: ", zap.Error(err))
		return data, err
	}

	loadFrom, err := os.Open(dataFile)

	if err != nil {
		kp.log.Info("Empty key/value store!: ", zap.Error(err))
		return data, err
	}
	defer loadFrom.Close()

	decoder := json.NewDecoder(loadFrom)
	for decoder.More() {
		var m models.DataUser
		err := decoder.Decode(&m)
		data[m.Email] = m

		if err != nil {
			kp.log.Info("cannot decode JSON file: ", zap.Error(err))
		}
	}

	return data, nil
}

func (kp *FileKeeper) Save(key string, data models.DataURL) (models.DataURL, error) {

	dataFile := kp.path()
	var (
		action string
		err    error
		cfile  *os.File
	)

	if _, err = os.Stat(dataFile); err == nil {
		//file exists. Open file
		cfile, err = os.OpenFile(dataFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		action = "open"
	} else {
		//file not exists. Create file
		cfile, err = os.Create(dataFile)
		action = "create"
	}

	if err != nil {
		kp.log.Info("Cannot %v file: %v", zap.Any(action, zapcore.StringType), zap.Error(err))
		return data, err
	}
	defer cfile.Close()

	// check if there is an entry in the file
	if action == "open" {
		decoder := json.NewDecoder(cfile)
		for decoder.More() {
			var m models.DataURL
			err := decoder.Decode(&m)
			if err != nil {
				kp.log.Info("cannot decode JSON file: ", zap.Error(err))
			}
			if m.ShortURL == key {
				return m, storage.ErrConflict
			}
		}
	}

	var id string
	if data.UUID == "" {
		neuuid := uuid.New()
		id = neuuid.String()
	} else {
		id = data.UUID
	}

	du := models.DataURL{
		UUID: id, ShortURL: data.ShortURL,
		OriginalURL: data.OriginalURL}

	encoder := json.NewEncoder(cfile)
	err = encoder.Encode(du)
	if err != nil {
		kp.log.Info("cannot encode JSON data", zap.Error(err))
		return data, err
	}

	return du, nil
}

func (kp *FileKeeper) SaveUser(key string, data models.DataUser) (models.DataUser, error) {

	dataFile := kp.path()
	var (
		action string
		err    error
		cfile  *os.File
	)

	if _, err = os.Stat(dataFile); err == nil {
		//file exists. Open file
		cfile, err = os.OpenFile(dataFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		action = "open"
	} else {
		//file not exists. Create file
		cfile, err = os.Create(dataFile)
		action = "create"
	}

	if err != nil {
		kp.log.Info("Cannot %v file: %v", zap.Any(action, zapcore.StringType), zap.Error(err))
		return data, err
	}
	defer cfile.Close()

	// check if there is an entry in the file
	if action == "open" {
		decoder := json.NewDecoder(cfile)
		for decoder.More() {
			var m models.DataUser
			err := decoder.Decode(&m)
			if err != nil {
				kp.log.Info("cannot decode JSON file: ", zap.Error(err))
			}
			if m.Email == key {
				return m, storage.ErrConflict
			}
		}
	}

	var id string
	if data.UUID == "" {
		neuuid := uuid.New()
		id = neuuid.String()
	} else {
		id = data.UUID
	}

	du := models.DataUser{
		UUID: id, Email: data.Email,
		Hash: data.Hash, Name: data.Name}

	encoder := json.NewEncoder(cfile)
	err = encoder.Encode(du)
	if err != nil {
		kp.log.Info("cannot encode JSON data", zap.Error(err))
		return data, err
	}

	return du, nil
}

func (kp *FileKeeper) SaveBatch(data storage.StorageURL) error {

	dataFile := kp.path()
	var (
		action string
		err    error
		cfile  *os.File
	)

	if _, err = os.Stat(dataFile); err == nil {
		//file exists. Open file
		cfile, err = os.OpenFile(dataFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		action = "open"
	} else {
		//file not exists. Create file
		cfile, err = os.Create(dataFile)
		action = "create"
	}

	if err != nil {
		kp.log.Info("Cannot %v file: %v", zap.Any(action, zapcore.StringType), zap.Error(err))
		return err
	}
	defer cfile.Close()

	for k, v := range data {

		var id string
		if v.UUID == "" {
			neuuid := uuid.New()
			id = neuuid.String()
		} else {
			id = v.UUID
		}

		du := models.DataURL{
			UUID: id, ShortURL: k,
			OriginalURL: v.OriginalURL}
		encoder := json.NewEncoder(cfile)
		err = encoder.Encode(du)
		if err != nil {
			kp.log.Info("cannot encode JSON data", zap.Error(err))
			return err
		}
	}

	return nil
}

func (kp *FileKeeper) Ping() bool { return true }

func (kp *FileKeeper) Close() bool { return true }
