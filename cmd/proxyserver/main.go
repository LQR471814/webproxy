package main

import (
	_ "embed"
	"log"
	"net"
	"net/http"
	"webproxy/pkg/server"
)

type ContentHandler struct {
	Contents    []byte
	ContentType string
	Location    string
	Redirect    http.Handler
}

func (h ContentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == h.Location && r.URL.RawQuery == "" {
		w.Header().Set("Content-Type", h.ContentType)
		w.Write(h.Contents)
		return
	}
	h.Redirect.ServeHTTP(w, r)
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("%v %v failed due to %v", r.Method, r.URL, err)
		}
	}()
}

//go:embed web/src/home/index.html
var homepage []byte

func main() {
	s := http.Server{Handler: ContentHandler{
		Location:    "/",
		Contents:    homepage,
		ContentType: "text/html",
		Redirect:    server.NewProxyHandler(true),
	}}

	l, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on port 3000")
	s.Serve(l)
}
