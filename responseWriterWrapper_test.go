package filter

import (
	"bytes"
	. "gopkg.in/check.v1"
	"net/http"
	"reflect"
)

type ResponseWriterWrapperTest struct{}

func init() {
	Suite(&ResponseWriterWrapperTest{})
}

func (s *ResponseWriterWrapperTest) Test_newResponseWriterWrapperFor(c *C) {
	original := newMockResponseWriter()
	beforeFirstWrite := func(*responseWriterWrapper) bool {
		return false
	}
	wrapper := newResponseWriterWrapperFor(original, beforeFirstWrite)
	c.Assert(wrapper.delegate, DeepEquals, original)
	c.Assert(wrapper.buffer, IsNil)
	c.Assert(reflect.ValueOf(wrapper.beforeFirstWrite), Equals, reflect.ValueOf(beforeFirstWrite))
	c.Assert(wrapper.bodyAllowed, Equals, true)
	c.Assert(wrapper.firstContentWritten, Equals, false)
}

func (s *ResponseWriterWrapperTest) Test_Header(c *C) {
	original := newMockResponseWriter()
	original.header.Add("a", "1")
	original.header.Add("b", "2")
	wrapper := newResponseWriterWrapperFor(original, nil)

	c.Assert(wrapper.Header().Get("a"), Equals, "1")
	c.Assert(wrapper.Header().Get("b"), Equals, "2")
	c.Assert(wrapper.Header().Get("c"), Equals, "")

	wrapper.Header().Del("a")
	wrapper.Header().Add("c", "3")
	c.Assert(original.header.Get("a"), Equals, "")
	c.Assert(original.header.Get("c"), Equals, "3")
}

func (s *ResponseWriterWrapperTest) Test_WriteHeader(c *C) {
	original := newMockResponseWriter()
	wrapper := newResponseWriterWrapperFor(original, nil)

	c.Assert(original.status, Equals, 0)

	wrapper.WriteHeader(200)
	c.Assert(original.status, Equals, 200)
	c.Assert(wrapper.bodyAllowed, Equals, true)

	wrapper.WriteHeader(204)
	c.Assert(original.status, Equals, 204)
	c.Assert(wrapper.bodyAllowed, Equals, false)
}

func (s *ResponseWriterWrapperTest) Test_bodyAllowedForStatus(c *C) {
	c.Assert(bodyAllowedForStatus(200), Equals, true)
	c.Assert(bodyAllowedForStatus(208), Equals, true)
	c.Assert(bodyAllowedForStatus(404), Equals, true)
	c.Assert(bodyAllowedForStatus(500), Equals, true)
	c.Assert(bodyAllowedForStatus(503), Equals, true)
	for i := 100; i < 200; i++ {
		c.Assert(bodyAllowedForStatus(i), Equals, false)
	}
	c.Assert(bodyAllowedForStatus(204), Equals, false)
	c.Assert(bodyAllowedForStatus(304), Equals, false)
}

func (s *ResponseWriterWrapperTest) Test_WriteWithoutRecording(c *C) {
	beforeFirstWriteCalled := false
	original := newMockResponseWriter()
	beforeFirstWrite := func(*responseWriterWrapper) bool {
		beforeFirstWriteCalled = true
		return false
	}
	wrapper := newResponseWriterWrapperFor(original, beforeFirstWrite)
	len, err := wrapper.Write([]byte(""))
	c.Assert(len, Equals, 0)
	c.Assert(err, IsNil)
	c.Assert(wrapper.firstContentWritten, Equals, false)

	len, err = wrapper.Write([]byte("foo"))
	c.Assert(len, Equals, 3)
	c.Assert(err, IsNil)
	c.Assert(wrapper.firstContentWritten, Equals, true)

	len, err = wrapper.Write([]byte("bar"))
	c.Assert(len, Equals, 3)
	c.Assert(err, IsNil)

	c.Assert(original.buffer.Bytes(), DeepEquals, []byte("foobar"))
	c.Assert(wrapper.buffer, IsNil)
	c.Assert(wrapper.firstContentWritten, Equals, true)
	c.Assert(wrapper.isBodyAllowed(), Equals, true)
	c.Assert(wrapper.recorded(), DeepEquals, []byte{})
	c.Assert(wrapper.wasSomethingRecorded(), Equals, false)
}

func (s *ResponseWriterWrapperTest) Test_WriteWithRecording(c *C) {
	beforeFirstWriteCalled := false
	original := newMockResponseWriter()
	beforeFirstWrite := func(*responseWriterWrapper) bool {
		beforeFirstWriteCalled = true
		return true
	}
	wrapper := newResponseWriterWrapperFor(original, beforeFirstWrite)
	len, err := wrapper.Write([]byte(""))
	c.Assert(len, Equals, 0)
	c.Assert(err, IsNil)
	c.Assert(wrapper.firstContentWritten, Equals, false)

	len, err = wrapper.Write([]byte("foo"))
	c.Assert(len, Equals, 3)
	c.Assert(err, IsNil)
	c.Assert(wrapper.firstContentWritten, Equals, true)

	len, err = wrapper.Write([]byte("bar"))
	c.Assert(len, Equals, 3)
	c.Assert(err, IsNil)

	c.Assert(original.buffer.Bytes(), DeepEquals, []byte(nil))
	c.Assert(wrapper.buffer.Bytes(), DeepEquals, []byte("foobar"))
	c.Assert(wrapper.firstContentWritten, Equals, true)
	c.Assert(wrapper.isBodyAllowed(), Equals, true)
	c.Assert(wrapper.recorded(), DeepEquals, []byte("foobar"))
	c.Assert(wrapper.wasSomethingRecorded(), Equals, true)
}

///////////////////////////////////////////////////////////////////////////////////////////
// MOCKS
///////////////////////////////////////////////////////////////////////////////////////////

func newMockResponseWriter() *mockResponseWriter {
	result := new(mockResponseWriter)
	result.header = http.Header{}
	result.buffer = new(bytes.Buffer)
	return result
}

type mockResponseWriter struct {
	header http.Header
	status int
	buffer *bytes.Buffer
}

func (instance *mockResponseWriter) Header() http.Header {
	return instance.header
}

func (instance *mockResponseWriter) WriteHeader(status int) {
	instance.status = status
}

func (instance *mockResponseWriter) Write(content []byte) (int, error) {
	return instance.buffer.Write(content)
}
