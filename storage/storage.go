package storage

type Storage interface {
	Get(path string) ([]byte, error)
	Set(path string, data []byte) error
}
