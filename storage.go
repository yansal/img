package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type storage interface {
	Get(path string) ([]byte, error)
	Set(path string, data []byte) error
}

type local struct{ Base string }

func (l *local) Get(path string) ([]byte, error) {
	abs := filepath.Join(l.Base, path)
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil, nil
	}
	return ioutil.ReadFile(abs)
}

func (l *local) Set(path string, data []byte) error {
	return ioutil.WriteFile(filepath.Join(l.Base, path), data, 0644)
}
