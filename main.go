package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
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
	log.Fatal(http.ListenAndServe(":"+port, &handler{}))
}

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.serveHTTP(w, r)
	if err == nil {
		return
	}

	herr, ok := err.(httpError)
	if !ok {
		herr = httpError{err: err, code: http.StatusInternalServerError}
	}
	http.Error(w, fmt.Sprintf("%+v", herr.Error()), herr.code)
}

type httpError struct {
	err  error
	code int
}

func (e httpError) Error() string { return e.err.Error() }

func (h *handler) serveHTTP(w http.ResponseWriter, r *http.Request) error {
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
	path, url     string
	width, height int
}

func bind(r *http.Request) (payload, error) {
	var p payload

	p.path = r.FormValue("path")
	p.url = r.FormValue("url")
	if p.path == "" && p.url == "" {
		return p, httpError{err: errors.New("path or url are required"), code: http.StatusBadRequest}
	} else if p.path != "" && p.url != "" {
		return p, httpError{err: errors.New("only one of path and url must be present"), code: http.StatusBadRequest}
	}
	// TODO: validate url?

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
	var (
		rc  io.ReadCloser
		err error
	)
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
	return buf.Bytes(), err
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
