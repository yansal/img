package manager

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/yansal/img/img"
	"github.com/yansal/img/storage"
)

type Manager struct {
	storage storage.Storage
}

func New(storage storage.Storage) *Manager {
	return &Manager{storage: storage}
}

func (m *Manager) Process(ctx context.Context, payload Payload) ([]byte, error) {
	b, err := m.get(ctx, payload)
	if err != nil {
		return nil, err
	}
	return img.Process(b, img.Option{Width: payload.Width, Height: payload.Height})
}

type Payload struct {
	Path, URL     string
	Width, Height int
}

func (m *Manager) get(ctx context.Context, payload Payload) ([]byte, error) {
	if payload.Path != "" {
		return m.storage.Get(ctx, payload.Path)
	} else if payload.URL != "" {
		return m.getURL(ctx, payload.URL)
	}
	return nil, errors.New("one of path or url must be set")
}

func (m *Manager) getURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
