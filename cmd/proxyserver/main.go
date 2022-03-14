package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"webproxy/pkg/server"
)

type ContentHandler struct {
	Contents    []byte
	ContentType string
	Location    string
	Redirect    http.Handler
}

func (h ContentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == h.Location && r.URL.Query().Get("proxyTargetURI") == "" {
		w.Header().Set("Content-Type", h.ContentType)
		w.Write(h.Contents)
		return
	}
	h.Redirect.ServeHTTP(w, r)
}

func main() {
	f, err := os.Open("web/src/index.html")
	if err != nil {
		log.Fatal(err)
	}

	homepage, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	proxyHandler, err := server.NewProxyHandler()
	if err != nil {
		log.Fatal(err)
	}

	s := http.Server{Handler: ContentHandler{
		Location:    "/",
		Contents:    homepage,
		ContentType: "text/html",
		Redirect:    proxyHandler,
	}}

	l, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on port 3000")
	s.Serve(l)
}
