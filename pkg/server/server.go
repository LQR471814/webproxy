package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

const TARGET_QUERY_NAME = "proxyTargetURI"

type address = string
type domain = string

type ProxyHandler struct {
	Inject   string                        //inject javascript into site
	Sessions map[string]map[string]Session //map[address]map[domain]Session
}

func (h ProxyHandler) Session(address, domain string) Session {
	if _, ok := h.Sessions[address]; !ok {
		h.Sessions[address] = make(map[string]Session)
	}
	if _, ok := h.Sessions[address][domain]; !ok {
		h.Sessions[address][domain] = Session{
			Client:       &http.Client{},
			TargetDomain: domain,
		}
	}
	return h.Sessions[address][domain]
}

func (h ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//? Get request url
	addr := r.RemoteAddr

	target := r.URL.Query().Get(TARGET_QUERY_NAME)
	if len(target) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Missing proxyTargetURI!"))
		return
	}

	targetURL, err := NormalizeURL(target)
	if err != nil {
		panic(err)
	}

	log.Printf("Requested %v\n", targetURL)

	//? Initialize and send request
	patchedRequest := *r
	patchedRequest.Host = ""
	patchedRequest.RequestURI = ""
	patchedRequest.URL = targetURL
	patchedRequest.Header.Set("Accept-Encoding", "gzip")

	response, err := h.Session(addr, targetURL.Host).Client.Do(&patchedRequest)
	if err != nil {
		panic(err)
	}

	//? Decompress and decode response body
	var decompressed io.Reader
	if response.Header.Get("Content-Encoding") == "gzip" {
		decompressed, err = gzip.NewReader(response.Body)
		if err != nil {
			panic(err)
		}
	} else {
		decompressed = response.Body
	}

	decoded, err := DecodeHTMLBody(decompressed, "utf-8")
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(decoded)
	if err != nil {
		panic(err)
	}

	urlREPL := func(target []byte) []byte {
		parsed := &url.URL{}
		parsed.Parse(string(target))
		return []byte((&url.URL{
			Scheme:   parsed.Scheme,
			Host:     r.URL.Host,
			RawQuery: TARGET_QUERY_NAME + "=" + url.QueryEscape(parsed.Host),
		}).String())
	}

	mimetype := strings.Split(response.Header.Get("Content-Type"), ";")[0]

	//? Handle different data types differently
	switch mimetype {
	case "text/html":
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			panic(err)
		}

		buff := new(bytes.Buffer)
		err = goquery.Render(buff, doc.Selection)
		if err != nil {
			panic(err)
		}

		doc.Find("body").AppendNodes(&html.Node{
			Type: html.ElementNode,
			Data: "script",
			FirstChild: &html.Node{
				Type: html.RawNode,
				Data: strings.Replace(
					h.Inject, "${TARGET_DOMAIN}",
					targetURL.Host, 1,
				),
			},
		})

		buff = new(bytes.Buffer)
		err = goquery.Render(buff, doc.Selection)
		if err != nil {
			panic(err)
		}
		body = buff.Bytes()
	case "text/css":
		for _, match := range CSSUrlMatch.FindAllSubmatchIndex(body, -1) {
			result := urlREPL(body[match[2]:match[3]])
			body = append(body[:match[2]], append(result, body[match[3]:]...)...)
		}
	case "text/ecmascript", "text/javascript", "text/markdown", "text/xml":
		body = StrictUrlMatch.ReplaceAllFunc(body, urlREPL)
	}

	//? Header shenanigans
	CopyHeaders(w.Header(), response.Header)

	//? Deal with encodings
	w.Header().Set("Content-Encoding", "gzip")

	buff := new(bytes.Buffer)

	encodedBody := gzip.NewWriter(buff)
	_, err = encodedBody.Write(body)
	if err != nil {
		panic(err)
	}
	err = encodedBody.Close()
	if err != nil {
		panic(err)
	}

	body = buff.Bytes()

	//? Fix content length and send response
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	_, err = w.Write(body)
	if err != nil {
		panic(err)
	}
}

func NewProxyHandler() (ProxyHandler, error) {
	f, err := os.Open("web/dist/inject.min.js")
	if err != nil {
		return ProxyHandler{}, err
	}
	script, err := io.ReadAll(f)
	if err != nil {
		return ProxyHandler{}, err
	}
	return ProxyHandler{
		Inject:   string(script),
		Sessions: make(map[string]map[string]Session),
	}, nil
}
