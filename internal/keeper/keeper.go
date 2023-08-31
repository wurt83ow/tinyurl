package keeper

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type DataURL struct {
	UUID        int64  `json:"result"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Log interface {
	Info(string, ...zapcore.Field)
}

type Keeper struct {
	data map[string]string
	path func() string
	log  Log
}

func NewKeeper(path func() string, log Log) *Keeper {
	return &Keeper{
		data: make(map[string]string),
		path: path,
		log:  log,
	}
}

func (kp *Keeper) Load() error {
	dataFile := kp.path()
	if _, err := os.Stat(dataFile); err != nil {
		kp.log.Info("file not found", zap.Error(err))
		return nil
	}

	loadFrom, err := os.Open(dataFile)

	if err != nil {
		kp.log.Info("Empty key/value store!", zap.Error(err))
		return err
	}
	defer loadFrom.Close()

	decoder := json.NewDecoder(loadFrom)
	for decoder.More() {
		var m DataURL
		err := decoder.Decode(&m)
		kp.data[m.ShortURL] = m.OriginalURL

		if err != nil {
			kp.log.Info("cannot decode JSON file", zap.Error(err))
		}
	}

	return nil
}

func (kp *Keeper) Save(data map[string]string) error {

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
		data := DataURL{
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
