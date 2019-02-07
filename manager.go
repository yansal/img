package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
)

type manager struct{ storage storage }

func (m *manager) resizeImage(p payload) ([]byte, error) {
	b, err := m.storage.Get(p.hash())
	if err != nil {
		return nil, err
	} else if b != nil {
		return b, nil
	}

	var rc io.ReadCloser
	if p.path != "" {
		rc, err = getpath(p.path)
	} else if p.url != "" {
		rc, err = geturl(p.url)
	}
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	img, format, err := image.Decode(rc)
	if err != nil {
		return nil, err
	}

	if p.width != 0 || p.height != 0 {
		img = imaging.Resize(img, p.width, p.height, imaging.Lanczos)
	}

	var buf bytes.Buffer
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
	if err != nil {
		return nil, err
	}

	b = buf.Bytes()
	return b, m.storage.Set(p.hash(), b)
}

func getpath(s string) (*os.File, error) {
	return os.Open(s)
}

func geturl(s string) (io.ReadCloser, error) {
	resp, err := http.Get(s)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
