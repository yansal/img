package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/yansal/img/storage/backends/local"
	"github.com/yansal/img/storage/backends/s3"
)

func main() {
	m := manager{cache: os.Getenv("NOCACHE") == ""}

	bucket := os.Getenv("S3BUCKET")
	if bucket != "" {
		var err error
		m.storage, err = s3.New(bucket)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		m.storage = &local.Storage{}
	}

	http.Handle("/", &handler{m: m})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type handler struct{ m manager }

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

	img, err := h.m.process(r.Context(), payload)
	if err != nil {
		return err
	}

	w.Write(img)
	return nil
}

type payload struct {
	path, url     string
	width, height int
	cache         bool
}

func (p payload) hash() string {
	return hash(p.path + p.url + strconv.Itoa(p.width) + strconv.Itoa(p.height))
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

	p.cache = r.FormValue("nocache") == ""

	return p, nil
}

func hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
