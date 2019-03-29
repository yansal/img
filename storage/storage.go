package storage

import "context"

type Storage interface {
	Get(ctx context.Context, path string) ([]byte, error)
}
