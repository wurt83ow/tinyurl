package filekeeper

import (
	"encoding/json"
	"os"

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
	return &FileKeeper{
		path: path,
		log:  log,
	}
}

func (kp *FileKeeper) Load() (map[string]string, error) {
	dataFile := kp.path()
	data := make(map[string]string)

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
		data[m.ShortURL] = m.OriginalURL

		if err != nil {
			kp.log.Info("cannot decode JSON file", zap.Error(err))
		}
	}

	return data, nil
}

func (kp *FileKeeper) Save(data map[string]string) error {

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

	var i int64 = 0

	for k, v := range data {
		i++
		data := models.DataURL{
			UUID: i, ShortURL: k,
			OriginalURL: v}
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
