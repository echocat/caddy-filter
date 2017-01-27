package filter

import (
	"bytes"
	"errors"
	"net/http"
)

func newResponseWriterWrapperFor(delegate http.ResponseWriter, beforeFirstWrite func(*responseWriterWrapper) bool) *responseWriterWrapper {
	return &responseWriterWrapper{
		delegate:            delegate,
		beforeFirstWrite:    beforeFirstWrite,
		statusSetAtDelegate: 200,
		bodyAllowed:         true,
		maximumBufferSize:   -1,
		header:              delegate.Header(),
	}
}

type responseWriterWrapper struct {
	delegate            http.ResponseWriter
	buffer              *bytes.Buffer
	beforeFirstWrite    func(*responseWriterWrapper) bool
	bodyAllowed         bool
	firstContentWritten bool
	headerSetAtDelegate bool
	statusSetAtDelegate int
	maximumBufferSize   int
	header              http.Header
}

func (instance *responseWriterWrapper) Header() http.Header {
	return instance.header
}

func (instance *responseWriterWrapper) WriteHeader(status int) {
	instance.bodyAllowed = bodyAllowedForStatus(status)
	instance.statusSetAtDelegate = status
}

func (instance *responseWriterWrapper) Write(content []byte) (int, error) {
	if len(content) <= 0 {
		return 0, nil
	}

	if !instance.firstContentWritten {
		if instance.beforeFirstWrite(instance) {
			instance.buffer = new(bytes.Buffer)
		} else {
			instance.buffer = nil
		}
		instance.firstContentWritten = true
	}

	if instance.buffer == nil {
		return instance.delegate.Write(content)
	}

	if (instance.maximumBufferSize >= 0) &&
		((instance.buffer.Len() + len(content)) > instance.maximumBufferSize) {
		_, err := instance.delegate.Write(instance.buffer.Bytes())
		if err != nil {
			return 0, err
		}
		instance.buffer = nil
		return instance.delegate.Write(content)
	}

	return instance.buffer.Write(content)
}

func (instance *responseWriterWrapper) writeToDelegate(content []byte) (int, error) {
	if !instance.headerSetAtDelegate {
		err := instance.writeHeadersToDelegate()
		if err != nil {
			return 0, err
		}
	}
	return instance.delegate.Write(content)
}

func (instance *responseWriterWrapper) writeHeadersToDelegate() error {
	if instance.headerSetAtDelegate {
		return errors.New("Headers already set at response.")
	}
	instance.delegate.WriteHeader(instance.statusSetAtDelegate)
	for key, values := range instance.header {
		for _, values := range values {
			instance.delegate.Header().Set(key, values)
		}
	}
	instance.headerSetAtDelegate = true
	return nil
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
