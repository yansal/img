package local

import (
	"context"
	"io/ioutil"
	"path/filepath"
)

type Storage struct{ Base string }

func (s *Storage) Get(ctx context.Context, path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(s.Base, path))
}
