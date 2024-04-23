package util

import (
	"errors"
	"io/fs"
	"os"
)

func CreateFileIfNotExists(filename string) error {
	f, err := os.Open(filename)
	f.Close()
	if errors.Is(err, fs.ErrNotExist) {
		f, err = os.Create(filename)
		f.Close()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
