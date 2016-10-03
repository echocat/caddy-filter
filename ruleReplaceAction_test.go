package filter

import (
	. "gopkg.in/check.v1"
	"net/http"
	"net/url"
	"regexp"
)

var (
	testUrl, _ = url.ParseRequestURI("http://foo.bar/my/path")
)

type ruleReplaceActionTest struct{}

func init() {
	Suite(&ruleReplaceActionTest{})
}

func (s *ruleReplaceActionTest) Test_replacer(c *C) {
	rra := &ruleReplaceAction{
		replacement:   []byte(""),
		searchPattern: nil,
		responseHeader: &http.Header{
			"A": []string{"foobar"},
		},
	}
	c.Assert(rra.replacer([]byte("My name is Caddy.")), DeepEquals, []byte("My name is Caddy."))
	rra.searchPattern = regexp.MustCompile("My name is (.*?)\\.")
	c.Assert(rra.replacer([]byte("My name is Caddy.")), DeepEquals, []byte(""))

	rra.replacement = []byte("Your name is {1}.")
	c.Assert(rra.replacer([]byte("My name is Caddy.")), DeepEquals, []byte("Your name is Caddy."))

	rra.replacement = []byte("Hi {1}! The header A is {response_header_A}.")
	c.Assert(rra.replacer([]byte("My name is Caddy.")), DeepEquals, []byte("Hi Caddy! The header A is foobar."))
}

func (s *ruleReplaceActionTest) Test_paramReplacer(c *C) {
	groups := [][]byte{
		[]byte("a"),
		[]byte("b"),
	}
	rra := &ruleReplaceAction{
		responseHeader: &http.Header{
			"A": []string{"c"},
		},
	}
	c.Assert(rra.paramReplacer([]byte("{0}"), groups), DeepEquals, []byte("a"))
	c.Assert(rra.paramReplacer([]byte("{1}"), groups), DeepEquals, []byte("b"))
	c.Assert(rra.paramReplacer([]byte("{response_header_A}"), groups), DeepEquals, []byte("c"))

	c.Assert(rra.paramReplacer([]byte(""), groups), DeepEquals, []byte(""))
	c.Assert(rra.paramReplacer([]byte("{}"), groups), DeepEquals, []byte("{}"))
	c.Assert(rra.paramReplacer([]byte("{2}"), groups), DeepEquals, []byte("{2}"))
	c.Assert(rra.paramReplacer([]byte("{response_headers_A}"), groups), DeepEquals, []byte("{response_headers_A}"))
	c.Assert(rra.paramReplacer([]byte("{foo}"), groups), DeepEquals, []byte("{foo}"))
}

func (s *ruleReplaceActionTest) Test_contextValueBy(c *C) {
	rra := &ruleReplaceAction{
		responseHeader: &http.Header{
			"A": []string{"fromResponse"},
		},
		request: &http.Request{
			Header: http.Header{
				"A": []string{"fromRequest"},
			},
		},
	}

	r, ok := rra.contextValueBy("request_header_A")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "fromRequest")

	r, ok = rra.contextValueBy("response_header_A")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "fromResponse")

	r, ok = rra.contextValueBy("foo")
	c.Assert(ok, Equals, false)
	c.Assert(r, Equals, "")
}

func (s *ruleReplaceActionTest) Test_contextRequestValueBy(c *C) {
	rra := &ruleReplaceAction{
		request: &http.Request{
			URL:        testUrl,
			Method:     "GET",
			Host:       "foo.bar",
			Proto:      "http",
			RemoteAddr: "1.2.3.4:6677",
			Header: http.Header{
				"A": []string{"1"},
				"B": []string{"2"},
			},
		},
	}
	r, ok := rra.contextRequestValueBy("header_A")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "1")

	r, ok = rra.contextRequestValueBy("header_B")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "2")

	r, ok = rra.contextRequestValueBy("header_C")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "")

	r, ok = rra.contextRequestValueBy("url")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, testUrl.String())

	r, ok = rra.contextRequestValueBy("path")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, testUrl.Path)

	r, ok = rra.contextRequestValueBy("method")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, rra.request.Method)

	r, ok = rra.contextRequestValueBy("host")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, rra.request.Host)

	r, ok = rra.contextRequestValueBy("proto")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, rra.request.Proto)

	r, ok = rra.contextRequestValueBy("remoteAddress")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, rra.request.RemoteAddr)

	r, ok = rra.contextRequestValueBy("headers_A")
	c.Assert(ok, Equals, false)
	c.Assert(r, Equals, "")

	r, ok = rra.contextRequestValueBy("foo")
	c.Assert(ok, Equals, false)
	c.Assert(r, Equals, "")
}

func (s *ruleReplaceActionTest) Test_contextResponseValueBy(c *C) {
	rra := &ruleReplaceAction{
		responseHeader: &http.Header{
			"A": []string{"1"},
			"B": []string{"2"},
		},
	}
	r, ok := rra.contextResponseValueBy("header_A")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "1")

	r, ok = rra.contextResponseValueBy("header_B")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "2")

	r, ok = rra.contextResponseValueBy("header_C")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "")

	r, ok = rra.contextResponseValueBy("headers_A")
	c.Assert(ok, Equals, false)
	c.Assert(r, Equals, "")

	r, ok = rra.contextResponseValueBy("foo")
	c.Assert(ok, Equals, false)
	c.Assert(r, Equals, "")
}
