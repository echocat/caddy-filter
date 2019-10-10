package filter

import (
	"fmt"
	. "gopkg.in/check.v1"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

const (
	testEnvironmentVariableName = "X_CADDY_FILTER_TESTING"
)

var (
	testUrl, _ = url.ParseRequestURI("http://foo.bar/my/path")
)

type ruleReplaceActionTest struct{}

func init() {
	Suite(&ruleReplaceActionTest{})
}

func (s *ruleReplaceActionTest) SetUpTest(c *C) {
	os.Setenv(testEnvironmentVariableName, c.TestName())
}

func (s *ruleReplaceActionTest) TearDownTest(c *C) {
	os.Unsetenv(testEnvironmentVariableName)
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
			"A":             []string{"c"},
			"Last-Modified": []string{"Tue, 01 Aug 2017 15:13:59 GMT"},
		},
	}
	yearString := time.Now().Format("2006-")

	c.Assert(rra.paramReplacer([]byte("{0}"), groups), DeepEquals, []byte("a"))
	c.Assert(rra.paramReplacer([]byte("{1}"), groups), DeepEquals, []byte("b"))
	c.Assert(rra.paramReplacer([]byte("{response_header_A}"), groups), DeepEquals, []byte("c"))

	c.Assert(rra.paramReplacer([]byte(""), groups), DeepEquals, []byte(""))
	c.Assert(rra.paramReplacer([]byte("{}"), groups), DeepEquals, []byte("{}"))
	c.Assert(rra.paramReplacer([]byte("{2}"), groups), DeepEquals, []byte("{2}"))
	c.Assert(rra.paramReplacer([]byte("{response_headers_A}"), groups), DeepEquals, []byte("{response_headers_A}"))
	c.Assert(rra.paramReplacer([]byte("{foo}"), groups), DeepEquals, []byte("{foo}"))
	c.Assert(rra.paramReplacer([]byte("{now:2006-}"), groups), DeepEquals, []byte(yearString))
	c.Assert(string(rra.paramReplacer([]byte("{response_header_last_modified}"), groups)), DeepEquals, "2017-08-01T15:13:59Z")
	c.Assert(string(rra.paramReplacer([]byte("{response_header_last_modified:RFC}"), groups)), DeepEquals, "2017-08-01T15:13:59Z")
	c.Assert(string(rra.paramReplacer([]byte("{response_header_last_modified:timestamp}"), groups)), DeepEquals, "1501600439000")
	c.Assert(string(rra.paramReplacer([]byte("{env_X_CADDY_FILTER_TESTING}"), groups)), DeepEquals, c.TestName())
}

func (s *ruleReplaceActionTest) Test_contextValueBy(c *C) {
	rra := &ruleReplaceAction{
		responseHeader: &http.Header{
			"A":             []string{"fromResponse"},
			"Last-Modified": []string{"Tue, 01 Aug 2017 15:13:59 GMT"},
		},
		request: &http.Request{
			Header: http.Header{
				"A": []string{"fromRequest"},
			},
		},
	}

	yearString := time.Now().Format("2006-")

	r, ok := rra.contextValueBy("request_header_A")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "fromRequest")

	r, ok = rra.contextValueBy("response_header_A")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "fromResponse")

	r, ok = rra.contextValueBy("now")
	c.Assert(ok, Equals, true)
	c.Assert(r, Matches, yearString+".*")

	r, ok = rra.contextValueBy("now:")
	c.Assert(ok, Equals, true)
	c.Assert(r, Matches, yearString+".*")

	r, ok = rra.contextValueBy("now:xxx2006-xxx")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, fmt.Sprintf("xxx%sxxx", yearString))

	r, ok = rra.contextValueBy("response_header_last_modified")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "2017-08-01T15:13:59Z")

	r, ok = rra.contextValueBy("response_header_last_modified:RFC")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "2017-08-01T15:13:59Z")

	r, ok = rra.contextValueBy("response_header_last_modified:timestamp")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "1501600439000")

	r, ok = rra.contextValueBy("env_X_CADDY_FILTER_TESTING")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, c.TestName())

	r, ok = rra.contextValueBy("env_X_CADDY_FILTER_TESTING-XYZ")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, "")

	r, ok = rra.contextValueBy("foo")
	c.Assert(ok, Equals, false)
	c.Assert(r, Equals, "")
}

func (s *ruleReplaceActionTest) Test_contextRequestValueBy(c *C) {
	rra := &ruleReplaceAction{
		request: &http.Request{
			URL:        testUrl,
			Method:     "GET",
			Scheme:       "https",
			Host:       "foo.bar",
			Proto:      "HTTP/2.0",
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

	r, ok = rra.contextRequestValueBy("scheme")
	c.Assert(ok, Equals, true)
	c.Assert(r, Equals, rra.request.Scheme)

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

func (s *ruleReplaceActionTest) Test_formatTimeBy(c *C) {
	rra := &ruleReplaceAction{}
	now, err := time.Parse(time.RFC3339Nano, "2017-08-15T14:00:00.123456789+02:00")
	c.Assert(err, IsNil)

	c.Assert(rra.formatTimeBy(now, ""), Equals, "2017-08-15T14:00:00+02:00")
	c.Assert(rra.formatTimeBy(now, "RFC"), Equals, "2017-08-15T14:00:00+02:00")
	c.Assert(rra.formatTimeBy(now, "RFC3339"), Equals, "2017-08-15T14:00:00+02:00")
	c.Assert(rra.formatTimeBy(now, "unix"), Equals, "1502798400")
	c.Assert(rra.formatTimeBy(now, "timestamp"), Equals, "1502798400123")
	c.Assert(rra.formatTimeBy(now, "2006-01-02"), Equals, "2017-08-15")
	c.Assert(rra.formatTimeBy(now, "xxx"), Equals, "xxx")
}
