package server

import (
	"fmt"
	"testing"
)

func errorMessage(testCase TestCase, got interface{}, err error) string {
	return fmt.Sprintf(
		"Test %v Got %v Expected %v Error %v",
		testCase.Test, got, testCase.Expect, err,
	)
}

type TestCase struct {
	Test   interface{}
	Expect interface{}
}

func TestPruneSlice(t *testing.T) {
	vectors := []TestCase{
		{[]string{"", "a", ""}, []string{"a"}},
		{[]string{"a"}, []string{"a"}},
		{[]string{"", ""}, []string{}},
	}

	for _, test := range vectors {
		result := PruneSlice(test.Test.([]string))
		for i, e := range test.Expect.([]string) {
			if e != result[i] {
				t.Error(errorMessage(test, result, nil))
			}
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	vectors := []TestCase{
		{"http://a.b.c", "http://a.b.c"},
		{"https://a.b.c/p/r.ext?q=s&qq=ss", "https://a.b.c/p/r.ext?q=s&qq=ss"},
		{"ws://a.b.c", "ws://a.b.c"},
		{"a.b.c", "http://a.b.c"},
	}

	for _, test := range vectors {
		result, err := NormalizeURL(test.Test.(string))
		onError := errorMessage(test, result, err)
		if err != nil || result.String() != test.Expect {
			t.Error(onError)
		}
	}
}

func TestResolveRelativeURL(t *testing.T) {
	vectors := []TestCase{
		{[]string{"https://a.b.c", "unknown.org"}, "https://a.b.c"},
		{[]string{"/gamer/moment", "unknown.org"}, "http://unknown.org/gamer/moment"},
		{[]string{"gamer/moment?q=s", "unknown.org"}, "http://unknown.org/gamer/moment?q=s"},
		{[]string{"../gamer/moment?q=s", "unknown.org"}, "http://unknown.org/../gamer/moment?q=s"},
	}

	for _, test := range vectors {
		result, err := ResolveRelativeURL(
			test.Test.([]string)[0],
			test.Test.([]string)[1],
		)

		onError := errorMessage(test, result, err)
		if err != nil || result != test.Expect {
			t.Error(onError)
		}
	}
}

func TestCSSURLReplacement(t *testing.T) {
	replaceWith := "http://target.com"
	vectors := []TestCase{
		{`url("http://google.com")`, `url("http://target.com")`},
		{`url('http://google.com')`, `url('http://target.com')`},
		{`url(http://google.com)`, `url(http://target.com)`},
		{`url(http://google.com); url(http://google.com)`, `url(http://target.com); url(http://target.com)`},
		{
			`.class{background-image: url("http://google.com/image.png"); width: 100px; height: 100px;}`,
			`.class{background-image: url("http://target.com"); width: 100px; height: 100px;}`,
		},
	}

	for _, test := range vectors {
		result := ReplaceCSSURLMatch(
			[]byte(test.Test.(string)),
			func(b []byte) []byte {
				return []byte(replaceWith)
			},
		)
		onError := errorMessage(test, string(result), nil)
		if string(result) != test.Expect {
			t.Error(onError)
		}
	}
}

func TestTransformURL(t *testing.T) {
	targetDomain := "target.com"
	vectors := []TestCase{
		{`https://www.external.com/path/to`, `http://www.proxy.com?proxyTargetURI=https%3A%2F%2Fwww.external.com%2Fpath%2Fto`},
		{`/relative/path`, `http://www.proxy.com?proxyTargetURI=http%3A%2F%2Ftarget.com%2Frelative%2Fpath`},
		{`../complex/path`, `http://www.proxy.com?proxyTargetURI=http%3A%2F%2Ftarget.com%2F..%2Fcomplex%2Fpath`},
	}

	for _, test := range vectors {
		result := TransformURL(test.Test.(string), "www.proxy.com", targetDomain)
		onError := errorMessage(test, string(result), nil)
		if string(result) != test.Expect {
			t.Error(onError)
		}
	}

}
