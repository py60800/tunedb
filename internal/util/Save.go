package util

import (
	"encoding/gob"
	"log"
	"os"
)

func GobSave(file string, data any) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Println("Save:", file, data)
	encoder := gob.NewEncoder(f)
	err = encoder.Encode(data)
	log.Println("Save:", file, data, err)
	return err
}
func GobLoad[T any](file string) (T, error) {
	var data T
	f, err := os.Open(file)
	if err != nil {
		return data, err
	}
	defer f.Close()
	decoder := gob.NewDecoder(f)
	err = decoder.Decode(&data)
	log.Println("Load:", file, data)
	return data, err
}
