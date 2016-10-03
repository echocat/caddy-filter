package filter

import (
	. "gopkg.in/check.v1"
	"net/http"
	"net/url"
	"regexp"
)

var (
	testUrl1, _ = url.ParseRequestURI("http://foo.bar/my/path.html")
	testUrl2, _ = url.ParseRequestURI("http://foo.bar/my/path.txt")
)

type ruleTest struct{}

func init() {
	Suite(&ruleTest{})
}

func (s *ruleTest) Test_matches_path(c *C) {
	req := &http.Request{}
	r := &rule{
		path: regexp.MustCompile(".*\\.html"),
	}

	req.URL = testUrl1
	c.Assert(r.matches(req, nil), Equals, true)

	req.URL = testUrl2
	c.Assert(r.matches(req, nil), Equals, false)
}

func (s *ruleTest) Test_matches_contentType(c *C) {
	header := http.Header{}
	r := &rule{
		contentType: regexp.MustCompile("text/html.*"),
	}

	header.Set("Content-Type", "text/html")
	c.Assert(r.matches(nil, &header), Equals, true)

	header.Set("Content-Type", "text/plain")
	c.Assert(r.matches(nil, &header), Equals, false)

	header.Del("Content-Type")
	c.Assert(r.matches(nil, &header), Equals, false)
}

func (s *ruleTest) Test_matches_combined(c *C) {
	req := &http.Request{}
	header := http.Header{}
	r := &rule{
		path:        regexp.MustCompile(".*\\.html"),
		contentType: regexp.MustCompile("text/html.*"),
	}

	req.URL = testUrl1
	header.Set("Content-Type", "text/html")
	c.Assert(r.matches(req, &header), Equals, true)

	req.URL = testUrl2
	header.Del("Content-Type")
	c.Assert(r.matches(nil, &header), Equals, false)
}

func (s *ruleTest) Test_execute(c *C) {
	req := &http.Request{}
	header := http.Header{}
	header.Set("Server", "Caddy")
	r := &rule{
		searchPattern: regexp.MustCompile("My name is (.*?)\\."),
		replacement:   []byte("Hi {1}! The name of this server is {response_header_Server}."),
	}

	result := r.execute(req, &header, []byte("Hello I'am a test.\nMy name is Test_execute."))
	c.Assert(string(result), Equals, "Hello I'am a test.\nHi Test_execute! The name of this server is Caddy.")

	r.searchPattern = nil
	result = r.execute(req, &header, []byte("foobar"))
	c.Assert(string(result), Equals, "foobar")
}
