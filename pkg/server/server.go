package server

import (
	"bytes"
	_ "embed"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

//go:embed post.min.js
var after []byte

//go:embed pre.min.js
var before []byte

var client *http.Client = &http.Client{}

const TARGET_QUERY_NAME = "proxyTargetURI"

type address = string
type domain = string

type Session struct {
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
	patchedRequest.Close = true

	//? Listen to redirects
	latestRedirect := session.TargetDomain
	redirects := 0
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
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
	response, err := client.Do(&patchedRequest)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	//? Account for redirects (in the case, the redirected domain contains
	//? a resource the original doesn't ex. google.com vs. www.google.com)
	session.TargetDomain = latestRedirect

	//? Decompress and decode response body
	urlREPL := func(target []byte) []byte {
		return []byte(TransformURL(
			string(target),
			r.URL.Host,
			session.TargetDomain,
		))
	}

	mimetype := strings.Split(response.Header.Get("Content-Type"), ";")[0]

	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	//? Handle different data types differently
	switch mimetype {
	case "text/html":
		body, err = OperateOnResponseBody(
			body,
			response.Header,
			func(b []byte) []byte {
				doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
				if err != nil {
					panic(err)
				}

				buff := new(bytes.Buffer)
				err = goquery.Render(buff, doc.Selection)
				if err != nil {
					panic(err)
				}

				fullTarget := targetURL.String()
				fillContext := func(script string) string {
					return strings.ReplaceAll(strings.ReplaceAll(
						string(script), "${TARGET_DOMAIN}",
						session.TargetDomain,
					), "${FULL_TARGET}", fullTarget)
				}

				doc.Find("head").BeforeNodes(&html.Node{
					Type: html.ElementNode,
					Data: "script",
					FirstChild: &html.Node{
						Type: html.RawNode,
						Data: fillContext(string(before)),
					},
				})

				doc.Find("body").AppendNodes(&html.Node{
					Type: html.ElementNode,
					Data: "script",
					FirstChild: &html.Node{
						Type: html.RawNode,
						Data: fillContext(string(after)),
					},
				})

				buff = new(bytes.Buffer)
				err = goquery.Render(buff, doc.Selection)
				if err != nil {
					panic(err)
				}
				return buff.Bytes()
			},
		)
		if err != nil {
			panic(err)
		}
	case "text/css":
		body, err = OperateOnResponseBody(
			body,
			response.Header,
			func(b []byte) []byte {
				return ReplaceCSSURLMatch(b, urlREPL)
			},
		)
		if err != nil {
			panic(err)
		}
	case "text/ecmascript", "text/javascript", "text/markdown", "text/xml":
		body, err = OperateOnResponseBody(
			body,
			response.Header,
			func(b []byte) []byte {
				return StrictUrlMatch.ReplaceAllFunc(b, urlREPL)
			},
		)
		if err != nil {
			panic(err)
		}
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

	//? Fix content length and send response
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	_, err = w.Write(body)
	if err != nil {
		panic(err)
	}
}

func NewProxyHandler(quiet bool) ProxyHandler {
	return ProxyHandler{
		Sessions: make(map[string]map[string]*Session),
		Quiet:    quiet,
	}
}
