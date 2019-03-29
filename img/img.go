package img

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

func NewProcessor(storage storage.Storage) *Processor {
	return &Processor{storage: storage}
}

type Processor struct{ storage storage.Storage }

func (p *Processor) Process(ctx context.Context, payload Payload) ([]byte, error) {
	ctx, task := trace.NewTask(ctx, "process")
	defer task.End()

	img, format, err := p.get(ctx, payload)
	if err != nil {
		return nil, err
	}
	b, err := p.resize(ctx, img, format, payload.Width, payload.Height)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (p *Processor) get(ctx context.Context, payload Payload) (image.Image, string, error) {
	defer trace.StartRegion(ctx, "get").End()

	var (
		b   []byte
		err error
	)
	if payload.Path != "" {
		b, err = p.storage.Get(payload.Path)
	} else if payload.URL != "" {
		b, err = p.geturl(ctx, payload.URL)
	}
	if err != nil {
		return nil, "", err
	}

	defer trace.StartRegion(ctx, "decode").End()
	return image.Decode(bytes.NewReader(b))
}

func (p *Processor) geturl(ctx context.Context, url string) ([]byte, error) {
	defer trace.StartRegion(ctx, "geturl").End()

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
	return b, nil
}

func (p *Processor) resize(ctx context.Context, img image.Image, format string, width, height int) ([]byte, error) {
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
