package filter

import (
	"errors"
	"fmt"
	. "gopkg.in/check.v1"
	"net/http"
	"regexp"
	"github.com/mholt/caddy/caddyhttp/fastcgi"
)

type filterTest struct {
	request     *http.Request
	writer      *mockResponseWriter
	nextHandler *mockHandler
	handler     *filterHandler
}

func init() {
	Suite(&filterTest{})
}

func (s *filterTest) SetUpTest(c *C) {
	s.request = &http.Request{
		URL: testUrl1,
	}
	s.writer = newMockResponseWriter()
	s.nextHandler = newMockHandler("Hello world!", 200)
	s.handler = &filterHandler{
		next: s.nextHandler,
		rules: []*rule{
			{
				path:          regexp.MustCompile(".*\\.html"),
				searchPattern: regexp.MustCompile("w(.)rld"),
				replacement:   []byte("2nd is '{1}'"),
			},
		},
		maximumBufferSize: defaultMaxBufferSize,
	}
}
func (s *filterTest) Test_withFiltering(c *C) {
	status, err := s.handler.ServeHTTP(s.writer, s.request)
	c.Assert(err, IsNil)
	c.Assert(status, Equals, 200)
	c.Assert(s.writer.buffer.String(), Equals, "Hello 2nd is 'o'!")
}

func (s *filterTest) Test_withBufferOverflow(c *C) {
	s.handler.maximumBufferSize = 5
	status, err := s.handler.ServeHTTP(s.writer, s.request)
	c.Assert(err, IsNil)
	c.Assert(status, Equals, 200)
	c.Assert(s.writer.buffer.String(), Equals, "Hello world!")
}

func (s *filterTest) Test_withoutFiltering(c *C) {
	s.request.URL = testUrl2
	status, err := s.handler.ServeHTTP(s.writer, s.request)
	c.Assert(err, IsNil)
	c.Assert(status, Equals, 200)
	c.Assert(s.writer.buffer.String(), Equals, "Hello world!")
}

func (s *filterTest) Test_withErrorInNext(c *C) {
	s.nextHandler.error = errors.New("Oops")
	status, err := s.handler.ServeHTTP(s.writer, s.request)
	c.Assert(err, DeepEquals, s.nextHandler.error)
	c.Assert(status, Equals, 200)
	c.Assert(s.writer.buffer.String(), Equals, "")
}

// This handles the bug https://github.com/echocat/caddy-filter/issues/4
// See filter.go for more details.
func (s *filterTest) Test_withLogErrorInNext(c *C) {
	s.nextHandler.error = fastcgi.LogError("Oops")
	status, err := s.handler.ServeHTTP(s.writer, s.request)
	c.Assert(err, DeepEquals, s.nextHandler.error)
	c.Assert(status, Equals, 200)
	c.Assert(s.writer.buffer.String(), Equals, "Hello 2nd is 'o'!")
}

func (s *filterTest) Test_withErrorInWriter(c *C) {
	s.writer.error = errors.New("Oops")
	status, err := s.handler.ServeHTTP(s.writer, s.request)
	c.Assert(err, DeepEquals, s.writer.error)
	c.Assert(status, Equals, 200)
	c.Assert(s.writer.buffer.String(), Equals, "")
}

///////////////////////////////////////////////////////////////////////////////////////////
// MOCKS
///////////////////////////////////////////////////////////////////////////////////////////

func newMockHandler(response string, status int) *mockHandler {
	result := new(mockHandler)
	result.response = response
	result.status = status
	return result
}

type mockHandler struct {
	response string
	status   int
	error    error
}

func (instance mockHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) (int, error) {
	writer.WriteHeader(instance.status)
	toReturn := []byte(instance.response)
	if len(toReturn) < 2 {
		return 0, fmt.Errorf("Response is too short: %v", toReturn)
	}
	middle := len(toReturn) / 2
	part1 := toReturn[:middle]
	part2 := toReturn[middle:]
	written, err := writer.Write(part1)
	if err != nil {
		return 0, err
	}
	if len(part1) != written {
		return 0, fmt.Errorf("Part 1 (%v) of response (%v) length was not written. Expected bytes written %v but got %v.", part1, toReturn, len(part1), written)
	}
	written, err = writer.Write(part2)
	if err != nil {
		return 0, err
	}
	if len(part2) != written {
		return 0, fmt.Errorf("Part 2 (%v) of response (%v) length was not written. Expected bytes written %v but got %v.", part2, toReturn, len(part2), written)
	}
	return instance.status, instance.error
}
