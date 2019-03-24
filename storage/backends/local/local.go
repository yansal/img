package local

import (
	"io/ioutil"
	"path/filepath"
)

type Storage struct{ Base string }

func (s *Storage) Get(path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(s.Base, path))
}
