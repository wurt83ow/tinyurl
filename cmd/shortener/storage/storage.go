package storage

import (
	"encoding/gob"
	"fmt"
	"os"
)

var (
	DATA     = make(map[string]string)
	DATAFILE = "/tmp/dataFile.gob"
)

func SaveURL(k string, v string) {
	if ADD(k, v) {
		err := save()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func save() error {
	removeFile(DATAFILE)

	saveTo, err := os.Create(DATAFILE)
	if err != nil {
		fmt.Println("Cannot create", DATAFILE)
		return err
	}
	defer saveTo.Close()

	encoder := gob.NewEncoder(saveTo)
	err = encoder.Encode(DATA)
	if err != nil {
		fmt.Println("Cannot save to", DATAFILE)
		return err
	}
	return nil
}

func Load() error {

	loadFrom, err := os.Open(DATAFILE)

	if err != nil {
		fmt.Println("Empty key/value store!")
		return err
	}
	defer loadFrom.Close()

	decoder := gob.NewDecoder(loadFrom)
	err = decoder.Decode(&DATA)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func ADD(k string, v string) bool {
	if k != "" && LOOKUP(k) == "" {
		DATA[k] = v
		return true
	}
	return false
}

func LOOKUP(k string) string {
	_, ok := DATA[k]
	if ok {
		return DATA[k]
	}
	return ""
}

func removeFile(path string) {
	if _, err := os.Stat(path); err == nil {
		err := os.Remove(path)
		if err != nil {
			fmt.Println(err)
		}
	}
}
