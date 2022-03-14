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
				t.Errorf(errorMessage(test, result, nil))
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
			t.Errorf(onError)
		}
	}
}

func TestResolveRelativeURL(t *testing.T) {
	vectors := []TestCase{
		{[]string{"https://a.b.c", "unknown.org"}, "https://a.b.c"},
		{[]string{"/gamer/moment", "unknown.org"}, "http://unknown.org/gamer/moment"},
		{[]string{"gamer/moment?q=s", "unknown.org"}, "http://unknown.org/gamer/moment?q=s"},
	}

	for _, test := range vectors {
		result, err := ResolveRelativeURL(
			test.Test.([]string)[0],
			test.Test.([]string)[1],
		)

		onError := errorMessage(test, result, err)
		if err != nil || result.String() != test.Expect {
			t.Errorf(onError)
		}
	}
}
