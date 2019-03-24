package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"

	"github.com/yansal/img/storage"
	"github.com/yansal/img/storage/backends/local"
	"github.com/yansal/img/storage/backends/s3"
	"golang.org/x/sync/semaphore"
)

func main() {
	var storage storage.Storage
	if bucket := os.Getenv("S3BUCKET"); bucket != "" {
		var err error
		storage, err = s3.New(bucket)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		storage = &local.Storage{}
	}

	http.Handle("/", &handler{
		m: &manager{storage: storage},
		s: semaphore.NewWeighted(int64(runtime.NumCPU())),
	})
	http.Handle("/favicon.ico", http.NotFoundHandler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type handler struct {
	m *manager
	s *semaphore.Weighted
}

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

	ctx := r.Context()
	h.s.Acquire(ctx, 1)
	defer h.s.Release(1)

	img, err := h.m.process(ctx, payload)
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
		width, err := strconv.Atoi(s)
		if err != nil {
			return p, httpError{err: err, code: http.StatusBadRequest}
		}
		p.width = width
	}

	s = r.FormValue("height")
	if s != "" {
		height, err := strconv.Atoi(s)
		if err != nil {
			return p, httpError{err: err, code: http.StatusBadRequest}
		}
		p.height = height
	}

	return p, nil
}
