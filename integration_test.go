package filter

import (
	"fmt"
	"github.com/echocat/caddy-filter/utils/test"
	. "github.com/echocat/gocheck-addons"
	_ "github.com/mholt/caddy/caddyhttp/errors"
	_ "github.com/mholt/caddy/caddyhttp/fastcgi"
	_ "github.com/mholt/caddy/caddyhttp/gzip"
	_ "github.com/mholt/caddy/caddyhttp/log"
	_ "github.com/mholt/caddy/caddyhttp/markdown"
	_ "github.com/mholt/caddy/caddyhttp/proxy"
	_ "github.com/mholt/caddy/caddyhttp/basicauth"
	_ "github.com/mholt/caddy/caddyhttp/root"
	. "gopkg.in/check.v1"
	"io"
	"io/ioutil"
	"net/http"
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

	etag := resp.Header.Get("Etag")
	c.Assert(etag, Not(Equals), "")

	resp, err = s.getWithEtag("http://localhost:22787/text.txt", etag)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 304)
	c.Assert(resp.ContentLength, Equals, int64(0))
	newEtag := resp.Header.Get("Etag")
	c.Assert(etag, Equals, newEtag)
}

func (s *integrationTest) Test_staticWithBasicAuth(c *C) {
	resp, err := http.Get("http://localhost:22790/text.txt")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "401 Unauthorized\n")
	c.Assert(resp.StatusCode, Equals, 401)
}

func (s *integrationTest) Test_staticWithGzip(c *C) {
	resp, err := http.Get("http://localhost:22788/text.txt")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello replaced world!\n")

	etag := resp.Header.Get("Etag")
	c.Assert(etag, Not(Equals), "")

	resp, err = s.getWithEtag("http://localhost:22788/text.txt", etag)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 304)
	c.Assert(resp.ContentLength, Equals, int64(0))
	newEtag := resp.Header.Get("Etag")
	c.Assert(etag, Equals, newEtag)
}

func (s *integrationTest) Test_proxy(c *C) {
	s.httpServer = test.NewTestingHttpServer(22775, false)

	resp, err := http.Get("http://localhost:22785/default")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello replaced world!")
}

func (s *integrationTest) Test_proxyWithGzip(c *C) {
	s.httpServer = test.NewTestingHttpServer(22776, false)

	resp, err := http.Get("http://localhost:22786/default")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, "Hello replaced world!")
}

func (s *integrationTest) Test_proxyWithGzipUpstream(c *C) {
	s.httpServer = test.NewTestingHttpServer(22777, true)

	resp, err := http.Get("http://localhost:22789/default")
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
	c.Assert(resp.Header.Get("Location"), Equals, "/another.cgi")
	c.Assert(string(content), Equals, "<a href=\"/another.cgi\">Moved Permanently</a>.")

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

func (s *integrationTest) getWithEtag(url string, etag string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("If-None-Match", etag)
	resp, err := http.DefaultClient.Do(req)
	return resp, err
}

