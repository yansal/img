package local

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type Storage struct{ Base string }

func (s *Storage) Get(path string) ([]byte, error) {
	abs := filepath.Join(s.Base, path)
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil, nil
	}
	return ioutil.ReadFile(abs)
}

func (s *Storage) Set(path string, data []byte) error {
	return ioutil.WriteFile(filepath.Join(s.Base, path), data, 0644)
}
