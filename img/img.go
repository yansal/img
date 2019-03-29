package img

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
)

type Option struct {
	Width, Height int
}

func Process(in []byte, option Option) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	if option.Width != 0 || option.Height != 0 {
		img = resize(img, option.Width, option.Height)
	}

	return encode(img, format)
}

func resize(img image.Image, width, height int) image.Image {
	return imaging.Resize(img, width, height, imaging.Lanczos)
}

func encode(img image.Image, format string) ([]byte, error) {
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
