package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/disintegration/imaging"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, handleError(handler)))
}

func handleError(h handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		herr, ok := err.(httpError)
		if !ok {
			log.Print(err)
			herr = httpError{err: err, code: http.StatusInternalServerError}
		}
		http.Error(w, herr.Error(), herr.code)
	}
}

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

type httpError struct {
	err  error
	code int
}

func (e httpError) Error() string { return e.err.Error() }

var handler = func(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	payload, err := bind(r)
	if err != nil {
		return err
	}

	img, err := process(payload)
	if err != nil {
		return err
	}
	w.Write(img)
	return nil
}

type payload struct {
	path          string
	width, height int
}

func bind(r *http.Request) (payload, error) {
	var p payload

	path := r.FormValue("path")
	if path == "" {
		return p, httpError{err: errors.New("path is required"), code: http.StatusBadRequest}
	}
	p.path = path

	s := r.FormValue("width")
	if s != "" {
		width, err := strconv.Atoi(r.FormValue("width"))
		if err != nil {
			return p, httpError{err: err, code: http.StatusBadRequest}
		}
		p.width = width
	}

	s = r.FormValue("height")
	if s != "" {
		height, err := strconv.Atoi(r.FormValue("height"))
		if err != nil {
			return p, httpError{err: err, code: http.StatusBadRequest}
		}
		p.height = height
	}

	return p, nil
}

func process(p payload) ([]byte, error) {
	f, err := os.Open(p.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, format, err := image.Decode(f)
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
	return buf.Bytes(), err
}
