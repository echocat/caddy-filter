package filter

import (
	. "github.com/echocat/gocheck-addons"
	. "gopkg.in/check.v1"

	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/echocat/caddy-filter/utils/test"
	_ "github.com/mholt/caddy/caddyhttp/errors"
	_ "github.com/mholt/caddy/caddyhttp/gzip"
	_ "github.com/mholt/caddy/caddyhttp/markdown"
	_ "github.com/mholt/caddy/caddyhttp/proxy"
	_ "github.com/mholt/caddy/caddyhttp/root"
	_ "github.com/mholt/caddy/caddyhttp/fastcgi"
	_ "github.com/mholt/caddy/caddyhttp/log"
	"io"
)

type integrationTest struct {
	httpServer *test.TestingHttpServer
	fcgiServer *test.TestingFcgiServer
	caddy      io.Closer
}

func init() {
	Suite(&integrationTest{})
}

func (s *integrationTest) Test_static(c *C) {
	resp, err := http.Get("http://localhost:22787/text.txt")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello replaced world!\n")
}

func (s *integrationTest) Test_staticWithGzip(c *C) {
	resp, err := http.Get("http://localhost:22788/text.txt")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello replaced world!\n")
}

func (s *integrationTest) Test_proxy(c *C) {
	s.httpServer = test.NewTestingHttpServer(22775)

	resp, err := http.Get("http://localhost:22785/default")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello replaced world!")
}

func (s *integrationTest) Test_proxyWithGzip(c *C) {
	s.httpServer = test.NewTestingHttpServer(22776)

	resp, err := http.Get("http://localhost:22786/default")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello replaced world!")
}

func (s *integrationTest) Test_fastcgi(c *C) {
	s.fcgiServer = test.NewTestingFcgiServer(22790)

	resp, err := http.Get("http://localhost:22780/index.cgi")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Contains, "<title>Hello replaced world!</title>")
}

func (s *integrationTest) Test_fastcgiWithGzip(c *C) {
	s.fcgiServer = test.NewTestingFcgiServer(22791)

	resp, err := http.Get("http://localhost:22781/index.cgi")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Contains, "<title>Hello replaced world!</title>")
}

func (s *integrationTest) Test_fastcgiWithRedirect(c *C) {
	s.fcgiServer = test.NewTestingFcgiServer(22792)

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", "http://localhost:22782/redirect.cgi", nil)

	c.Assert(err, IsNil)
	resp, err := client.Do(req)
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 301)
	c.Assert(resp.Status, Equals, "301 Moved Permanently")
	c.Assert(string(content), IsEmpty)

	resp2, err := http.Get("http://caddyserver.com")
	c.Assert(err, IsNil)
	defer resp2.Body.Close()
	content, err = ioutil.ReadAll(resp2.Body)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 301)
	c.Assert(string(content), Contains, "<title>Caddy - ")

	resp3, err := http.Get("http://localhost:22782/redirect.cgi")
	c.Assert(err, IsNil)
	defer resp3.Body.Close()
	content, err = ioutil.ReadAll(resp3.Body)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 301)
	c.Assert(string(content), Contains, "<title>Replaced another!</title>")
}

func (s *integrationTest) Test_markdown(c *C) {
	resp, err := http.Get("http://localhost:22783/index.md")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Contains, "<title>Hello replaced world!</title>")
}

func (s *integrationTest) Test_markdownWithGzip(c *C) {
	resp, err := http.Get("http://localhost:22784/index.md")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Contains, "<title>Hello replaced world!</title>")
}

func (s *integrationTest) SetUpTest(c *C) {
	c.Check(s.httpServer, IsNil)
	c.Check(s.fcgiServer, IsNil)
	c.Check(s.caddy, IsNil)
	s.caddy = test.NewTestingCaddy(fmt.Sprintf("%s.conf", c.TestName()))
}

func (s *integrationTest) TearDownTest(c *C) {
	if s.httpServer != nil {
		s.httpServer.Close()
	}
	if s.fcgiServer != nil {
		s.fcgiServer.Close()
	}
	if s.caddy != nil {
		s.caddy.Close()
	}
	s.httpServer = nil
	s.fcgiServer = nil
	s.caddy = nil
}
