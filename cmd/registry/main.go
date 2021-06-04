package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/imjasonh/image/pkg/registry"
)

var (
	port = flag.Int("port", 8080, "Port to listen on")
	path = flag.String("path", "layout", "Path to layout files")
)

func main() {
	flag.Parse()

	p, err := layout.FromPath(*path)
	if err != nil {
		log.Fatalf("layout.FromPath(%q): %v", *path, err)
	}

	http.Handle("/", registry.New(p))
	log.Printf("Serving on %d...", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
