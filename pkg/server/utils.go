package server

import (
	"bufio"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/htmlindex"
)

func PruneSlice(slice []string) []string {
	result := []string{}
	for _, e := range slice {
		if len(e) > 0 {
			result = append(result, e)
		}
	}
	return result
}

//This assumes a stricter form of url that must include a scheme
//if a scheme is omitted it assumes the entire url is a relative path
//This is a form of lenient url auto-completion that is supported within
//HTML, JS and CSS
func ResolveRelativeURL(target string, defaultDomain string) (*url.URL, error) {
	result := &url.URL{}
	result, err := result.Parse(target)
	if err != nil {
		return result, err
	}

	if len(result.Host) == 0 {
		result.Host = defaultDomain
		result.Scheme = "http"
	}

	return result, nil
}

//This assumes a more lenient form of url that behavior's best defined
//by the url search bar in a typical browser
//This function should only be used to mimic browser search bar behavior
//when accepting user input. JS, CSS and HTML prohibit these kinds of urls
func NormalizeURL(target string) (*url.URL, error) {
	result := &url.URL{}

	parsed, err := (&url.URL{}).Parse(target)
	if err != nil {
		return result, err
	}

	if len(parsed.Host) > 0 && len(parsed.Scheme) > 0 {
		result = parsed
	} else {
		if len(parsed.Host) == 0 && len(parsed.Scheme) == 0 {
			path := PruneSlice(strings.Split(parsed.Path, "/"))
			result.Host = path[0]
			if len(path) > 1 {
				result.Path = "/" + strings.Join(path[1:], "/")
			}
		}
		result.Scheme = "http"
	}

	return result, nil
}

func CopyHeaders(dst http.Header, headers http.Header) {
	for name, values := range headers {
		for _, v := range values {
			dst.Add(name, v)
		}
	}
}

var StrictUrlMatch = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
var CSSUrlMatch = regexp.MustCompile(`url\(['"]?([^'"]+)['"]?\)`)

func ReplaceCSSURLMatch(buff []byte, repl func([]byte) []byte) []byte {
	for _, match := range CSSUrlMatch.FindAllSubmatchIndex(buff, -1) {
		result := repl(buff[match[2]:match[3]])
		buff = append(buff[:match[2]], append(result, buff[match[3]:]...)...)
	}
	return buff
}

func DetectContentCharset(body io.Reader) string {
	r := bufio.NewReader(body)
	if data, err := r.Peek(1024); err == nil {
		if _, name, ok := charset.DetermineEncoding(data, ""); ok {
			return name
		}
	}
	return "utf-8"
}

// DecodeHTMLBody returns an decoding reader of the html Body for the specified `charset`
// If `charset` is empty, DecodeHTMLBody tries to guess the encoding from the content
func DecodeHTMLBody(body io.Reader, charset string) (io.Reader, error) {
	if charset == "" {
		charset = DetectContentCharset(body)
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, err
	}
	if name, _ := htmlindex.Name(e); name != "utf-8" {
		body = e.NewDecoder().Reader(body)
	}
	return body, nil
}
