package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"runtime/trace"

	"github.com/disintegration/imaging"
	"github.com/yansal/img/storage"
)

type manager struct {
	cache   bool
	storage storage.Storage
}

func (m *manager) process(ctx context.Context, p payload) ([]byte, error) {
	ctx, task := trace.NewTask(ctx, "process")
	defer task.End()

	if m.cache && p.cache {
		b, err := m.getcache(ctx, p.hash())
		if err != nil {
			return nil, err
		} else if b != nil {
			return b, nil
		}
	}

	img, format, err := m.get(ctx, p)
	if err != nil {
		return nil, err
	}

	b, err := m.resize(ctx, img, format, p.width, p.height)
	if err != nil {
		return nil, err
	}
	if m.cache && p.cache {
		return b, m.setcache(ctx, p.hash(), b)
	}
	return b, nil
}

func (m *manager) getcache(ctx context.Context, key string) ([]byte, error) {
	defer trace.StartRegion(ctx, "getcache").End()
	return m.storage.Get(key)
}

func (m *manager) setcache(ctx context.Context, key string, value []byte) error {
	defer trace.StartRegion(ctx, "setcache").End()
	return m.storage.Set(key, value)
}

func (m *manager) get(ctx context.Context, p payload) (image.Image, string, error) {
	defer trace.StartRegion(ctx, "get").End()

	var (
		b   []byte
		err error
	)
	if p.path != "" {
		b, err = ioutil.ReadFile(p.path)
	} else if p.url != "" {
		b, err = m.geturl(ctx, p.url, p.cache)
	}
	if err != nil {
		return nil, "", err
	}

	defer trace.StartRegion(ctx, "decode").End()
	return image.Decode(bytes.NewReader(b))
}

func (m *manager) geturl(ctx context.Context, url string, cache bool) ([]byte, error) {
	defer trace.StartRegion(ctx, "geturl").End()

	if m.cache && cache {
		b, err := m.getcache(ctx, hash(url))
		if err != nil {
			return nil, err
		} else if b != nil {
			return b, nil
		}
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if m.cache && cache {
		return b, m.setcache(ctx, hash(url), b)
	}
	return b, nil
}

func (m *manager) resize(ctx context.Context, img image.Image, format string, width, height int) ([]byte, error) {
	defer trace.StartRegion(ctx, "resize").End()

	if width != 0 || height != 0 {
		img = imaging.Resize(img, width, height, imaging.Lanczos)
	}

	defer trace.StartRegion(ctx, "encode").End()
	var (
		buf bytes.Buffer
		err error
	)
	switch format {
	case "gif":
		err = gif.Encode(&buf, img, nil)
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "png":
		err = png.Encode(&buf, img)
	default:
		return nil, fmt.Errorf("don't know how to encode format %s", format)
	}
	return buf.Bytes(), err
}
