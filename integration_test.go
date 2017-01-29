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
	resp, err := http.Get("http://localhost:22780/text.txt")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello replaced world!\n")
}

func (s *integrationTest) Test_staticWithGzip(c *C) {
	s.Test_static(c)
}

func (s *integrationTest) Test_proxy(c *C) {
	s.httpServer = test.NewTestingHttpServer(22770)

	resp, err := http.Get("http://localhost:22780/default")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello replaced world!")
}

func (s *integrationTest) Test_proxyWithGzip(c *C) {
	s.Test_proxy(c)
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
	s.Test_fastcgi(c)
}

func (s *integrationTest) Test_markdown(c *C) {
	resp, err := http.Get("http://localhost:22780/index.md")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Contains, "<title>Hello replaced world!</title>")
}

func (s *integrationTest) Test_markdownWithGzip(c *C) {
	s.Test_markdown(c)
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
