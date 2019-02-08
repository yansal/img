package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"

	"github.com/disintegration/imaging"
)

type manager struct{ storage storage }

func (m *manager) resizeImage(p payload) ([]byte, error) {
	if !p.nocache {
		b, err := m.storage.Get(p.hash())
		if err != nil {
			return nil, err
		} else if b != nil {
			return b, nil
		}
	}

	img, format, err := m.decodeImage(p)
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

	b := buf.Bytes()
	if !p.nocache {
		return b, m.storage.Set(p.hash(), b)
	}
	return b, nil
}

func (m *manager) decodeImage(p payload) (image.Image, string, error) {
	var (
		b   []byte
		err error
	)
	if p.path != "" {
		b, err = ioutil.ReadFile(p.path)
	} else if p.url != "" {
		b, err = m.geturl(p.url, p.nocache)
	}
	if err != nil {
		return nil, "", err
	}
	return image.Decode(bytes.NewReader(b))
}

func (m *manager) geturl(url string, nocache bool) ([]byte, error) {
	if !nocache {
		b, err := m.storage.Get(hash(url))
		if err != nil {
			return nil, err
		} else if b != nil {
			return b, nil
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if !nocache {
		return b, m.storage.Set(hash(url), b)
	}
	return b, nil
}
