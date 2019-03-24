package storage

type Storage interface {
	Get(path string) ([]byte, error)
}
