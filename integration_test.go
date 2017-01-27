package filter

import (
	"fmt"
	"github.com/echocat/caddy-filter/utils/test"
	. "github.com/echocat/gocheck-addons"
	"github.com/mholt/caddy"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"net/http"

	_ "github.com/mholt/caddy/caddyhttp/errors"
	_ "github.com/mholt/caddy/caddyhttp/gzip"
	_ "github.com/mholt/caddy/caddyhttp/markdown"
	_ "github.com/mholt/caddy/caddyhttp/proxy"
	_ "github.com/mholt/caddy/caddyhttp/root"
)

type integrationTest struct {
	server *test.TestingServer
	caddy  *caddy.Instance
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
	s.server = test.NewTestingServer(22770)

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
	c.Check(s.server, IsNil)
	s.caddy = test.NewTestingCaddy(fmt.Sprintf("%s.conf", c.TestName()))
}

func (s *integrationTest) TearDownTest(c *C) {
	if s.server != nil {
		s.server.Close()
	}
	if s.caddy != nil {
		s.caddy.Stop()
	}
	s.server = nil
	s.caddy = nil
}
