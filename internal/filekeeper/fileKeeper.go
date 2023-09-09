package filekeeper

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"
	"github.com/wurt83ow/tinyurl/internal/models"
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
		kp.log.Info("file not found", zap.Error(err))
		return data, err
	}

	loadFrom, err := os.Open(dataFile)

	if err != nil {
		kp.log.Info("Empty key/value store!", zap.Error(err))
		return data, err
	}
	defer loadFrom.Close()

	decoder := json.NewDecoder(loadFrom)
	for decoder.More() {
		var m models.DataURL
		err := decoder.Decode(&m)
		data[m.ShortURL] = m

		if err != nil {
			kp.log.Info("cannot decode JSON file", zap.Error(err))
		}
	}

	return data, nil
}

func (kp *FileKeeper) Save(data storage.StorageURL) error {
	fmt.Println("88888888888888888888888888888888888888888888888888")
	dataFile := kp.path()

	if _, err := os.Stat(dataFile); err == nil {
		err := os.Remove(dataFile)
		if err != nil {
			kp.log.Info("Cannot remove file", zap.Error(err))
		}
	}

	saveTo, err := os.Create(dataFile)
	if err != nil {
		kp.log.Info("Cannot create file", zap.Error(err))
		return err
	}
	defer saveTo.Close()

	for k, v := range data {
		var id string
		if v.UUID == "" {
			neuuid := uuid.New()
			id = neuuid.String()
		} else {
			id = v.UUID
		}

		data := models.DataURL{
			UUID: id, ShortURL: k,
			OriginalURL: v.OriginalURL}
		encoder := json.NewEncoder(saveTo)
		err = encoder.Encode(data)
		if err != nil {
			kp.log.Info("cannot encode JSON data", zap.Error(err))
			return err
		}
	}

	return nil
}

func (kp *FileKeeper) Ping() bool { return true }

func (kp *FileKeeper) Close() bool { return true }
