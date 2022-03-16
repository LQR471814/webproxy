package server

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

//go:embed inject.min.js
var injectScript []byte

const TARGET_QUERY_NAME = "proxyTargetURI"

type address = string
type domain = string

type Session struct {
	Client       *http.Client
	TargetDomain string
}

type ProxyHandler struct {
	Sessions map[string]map[string]*Session //map[address]map[domain]Session
	Quiet    bool
}

func (h ProxyHandler) Session(address, domain string) *Session {
	if _, ok := h.Sessions[address]; !ok {
		h.Sessions[address] = make(map[string]*Session)
	}
	if _, ok := h.Sessions[address][domain]; !ok {
		h.Sessions[address][domain] = &Session{
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
		targetDomainCookie, err := r.Cookie("Target-Domain")
		if err == nil {
			target = (&url.URL{
				Scheme:   "http",
				Host:     targetDomainCookie.Value,
				Path:     r.URL.Path,
				RawQuery: r.URL.RawQuery,
			}).String()
		} else {
			w.WriteHeader(400)
			w.Write([]byte("Missing proxyTargetURI!"))
			return
		}
	}

	targetURL, err := NormalizeURL(target)
	if err != nil {
		panic(err)
	}

	if !h.Quiet {
		log.Printf("Requested %v\n", targetURL)
	}

	//? Initialize client session
	session := h.Session(addr, targetURL.Host)

	//? Patch client request
	patchedRequest := *r
	patchedRequest.Host = ""
	patchedRequest.RequestURI = ""
	patchedRequest.URL = targetURL
	patchedRequest.Header.Set("Accept-Encoding", "gzip")

	//? Listen to redirects
	latestRedirect := session.TargetDomain
	redirects := 0
	session.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if redirects > 10 {
			return errors.New("amount of redirects exceeded 10")
		}
		if !h.Quiet {
			log.Println("  Redirected ->", req.URL)
		}
		latestRedirect = req.URL.Host
		redirects++
		return nil
	}

	//? Send request
	response, err := session.Client.Do(&patchedRequest)
	if err != nil {
		panic(err)
	}

	//? Account for redirects (in the case, the redirected domain contains
	//? a resource the original doesn't ex. google.com vs. www.google.com)
	session.TargetDomain = latestRedirect

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
		return []byte(TransformURL(
			string(target),
			r.URL.Host,
			session.TargetDomain,
		))
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
				Data: strings.ReplaceAll(
					string(injectScript), "${TARGET_DOMAIN}",
					session.TargetDomain,
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
		body = ReplaceCSSURLMatch(body, urlREPL)
	case "text/ecmascript", "text/javascript", "text/markdown", "text/xml":
		body = StrictUrlMatch.ReplaceAllFunc(body, urlREPL)
	}

	//? Copy target response headers
	CopyHeaders(w.Header(), response.Header)

	//? Add target domain as cookie in case a request
	//? whose url can't be affected by html or JS is requested.
	//?     Limitation: can only work with urls that aren't cross origin
	//?     so this method will not replace the url parameter
	w.Header().Add(
		"Set-Cookie",
		(&http.Cookie{
			Name:  "Target-Domain",
			Value: session.TargetDomain,
		}).String(),
	)

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

func NewProxyHandler() ProxyHandler {
	return ProxyHandler{
		Sessions: make(map[string]map[string]*Session),
		Quiet:    true,
	}
}
