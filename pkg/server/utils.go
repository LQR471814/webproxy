package server

import (
	"bufio"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/htmlindex"
)

//PruneSlice removes all empty strings from a slice
func PruneSlice(slice []string) []string {
	result := []string{}
	for _, e := range slice {
		if len(e) > 0 {
			result = append(result, e)
		}
	}
	return result
}

//ResolveRelativeURL will create an absolute url from a relative one
//This assumes a stricter form of url that must include a scheme
//if a scheme is omitted it assumes the entire url is a relative path
//This is a form of lenient url auto-completion that is supported within
//HTML, JS and CSS
func ResolveRelativeURL(target string, defaultDomain string) (string, error) {
	const length = 5 //1/26^5 (0.00000841%) chance of collision
	floor := math.Pow10(length)
	ceiling := math.Pow10(length+1) - 1
	escapeString := strconv.Itoa(int(floor + rand.Float64()*(ceiling-floor)))

	//? URL.Parse() will screw up periods in relative paths so they are
	//? escaped with a random string
	target = strings.ReplaceAll(target, ".", escapeString)

	result := &url.URL{}
	result, err := result.Parse(target)
	if err != nil {
		return "", err
	}

	if len(result.Host) == 0 {
		result.Host = defaultDomain
		result.Scheme = "http"
	}

	return strings.ReplaceAll(result.String(), escapeString, "."), nil
}

//NormalizeURL will attempt to normalize the more lenient form of url
//that behavior's best defined by the url search bar in a typical browser
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

//CopyHeaders copies the headers of 'headers' to 'dst'
func CopyHeaders(dst http.Header, headers http.Header) {
	for name, values := range headers {
		for _, v := range values {
			dst.Add(name, v)
		}
	}
}

//StrictUrlMatch matches urls that must contain a scheme
var StrictUrlMatch = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)

//CSSUrlMatch matches the CSS url() function and it's contents
var CSSUrlMatch = regexp.MustCompile(`url\(['"]?([^\)'"]+)['"]?\)`)

//ReplaceCSSURLMatch is a utility for replacing all the css url()
//function links in a given byteslice with the repl function
func ReplaceCSSURLMatch(buff []byte, repl func([]byte) []byte) []byte {
	offset := 0
	for _, match := range CSSUrlMatch.FindAllSubmatchIndex(buff, -1) {
		replace := repl(buff[match[2]+offset : match[3]+offset])
		buff = append(buff[:match[2]+offset], append(replace, buff[match[3]+offset:]...)...)
		offset += len(replace) - (match[3] - match[2])
	}
	return buff
}

//TransformURL performs the same function as it's javascript counterpart
//it takes the given url which may be a relative or absolute url and puts
//it into the query param while normalizing the url if it's relative
func TransformURL(target string, proxyDomain, targetDomain string) string {
	resolved, err := ResolveRelativeURL(
		target,
		targetDomain,
	)
	if err != nil {
		log.Println("WARN - During URL Transformation:", err)
		return target
	}
	return (&url.URL{
		Scheme:   "http",
		Host:     proxyDomain,
		RawQuery: TARGET_QUERY_NAME + "=" + url.QueryEscape(resolved),
	}).String()
}

//DetectContentCharset will attempt to detect the charset of a given reader
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
