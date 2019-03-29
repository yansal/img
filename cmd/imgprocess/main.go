package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/yansal/img/img"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("imgprocess: ")

	if isatty.IsTerminal(os.Stdout.Fd()) {
		log.Fatal("stdout must be redirected (you don't want to see binary data)")
	}
	height := flag.Int("height", 0, "resize image to height, don't resize if height are 0")
	width := flag.Int("width", 0, "resize image to width, don't resize if height are 0")
	flag.Parse()

	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	b, err := img.Process(in, img.Option{Height: *height, Width: *width})
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(b)
}
