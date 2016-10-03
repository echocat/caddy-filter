package filter

import (
	"bytes"
	"net/http"
)

func newResponseWriterWrapperFor(delegate http.ResponseWriter, beforeFirstWrite func(*responseWriterWrapper) bool) *responseWriterWrapper {
	return &responseWriterWrapper{
		delegate:          delegate,
		beforeFirstWrite:  beforeFirstWrite,
		bodyAllowed:       true,
		maximumBufferSize: -1,
	}
}

type responseWriterWrapper struct {
	delegate            http.ResponseWriter
	buffer              *bytes.Buffer
	beforeFirstWrite    func(*responseWriterWrapper) bool
	bodyAllowed         bool
	firstContentWritten bool
	maximumBufferSize   int
}

func (instance *responseWriterWrapper) Header() http.Header {
	return instance.delegate.Header()
}

func (instance *responseWriterWrapper) WriteHeader(status int) {
	instance.bodyAllowed = bodyAllowedForStatus(status)
	instance.delegate.WriteHeader(status)
}

func (instance *responseWriterWrapper) Write(content []byte) (int, error) {
	if len(content) <= 0 {
		return 0, nil
	}

	if !instance.firstContentWritten {
		if instance.beforeFirstWrite(instance) {
			instance.buffer = new(bytes.Buffer)
		}
		instance.firstContentWritten = true
	}

	if instance.maximumBufferSize >= 0 {
		if instance.buffer != nil {
			if (instance.buffer.Len() + len(content)) > instance.maximumBufferSize {
				_, err := instance.delegate.Write(instance.buffer.Bytes())
				if err != nil {
					return 0, err
				}
				instance.buffer = nil
				return instance.delegate.Write(content)
			}
		}
	}

	if instance.buffer != nil {
		return instance.buffer.Write(content)
	}

	return instance.delegate.Write(content)
}

func (instance *responseWriterWrapper) isBodyAllowed() bool {
	return instance.bodyAllowed
}

func (instance *responseWriterWrapper) wasSomethingRecorded() bool {
	return instance.buffer != nil && instance.buffer.Len() > 0
}

func (instance *responseWriterWrapper) recorded() []byte {
	buffer := instance.buffer
	if buffer == nil {
		return []byte{}
	}
	return buffer.Bytes()
}

func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}
