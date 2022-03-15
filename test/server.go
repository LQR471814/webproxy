package main

import (
	"log"
	"net"
	"net/http"
)

func main() {
	s := http.Server{
		Handler: http.FileServer(http.Dir("web")),
	}

	l, err := net.Listen("tcp", "127.0.0.1:5000")
	if err != nil {
		log.Fatal(err)
	}

	s.Serve(l)
}
