package filter

import (
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"net/http"
)

const defaultMaxBufferSize = 10 * 1024 * 1024

type filterHandler struct {
	next              httpserver.Handler
	rules             []*rule
	maximumBufferSize int
}

func (instance filterHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) (int, error) {
	header := writer.Header()
	wrapper := newResponseWriterWrapperFor(writer, func(wrapper *responseWriterWrapper) bool {
		for _, rule := range instance.rules {
			if rule.matches(request, &header) {
				return true
			}
		}
		return false
	})
	wrapper.maximumBufferSize = instance.maximumBufferSize
	result, err := instance.next.ServeHTTP(wrapper, request)
	if err != nil {
		return result, err
	}
	if !wrapper.wasSomethingRecorded() || !wrapper.isBodyAllowed() {
		return result, nil
	}
	body := wrapper.recorded()
	for _, rule := range instance.rules {
		if rule.matches(request, &header) {
			body = rule.execute(request, &header, body)
		}
	}
	_, err = writer.Write(body)
	if err != nil {
		return result, err
	}
	return result, nil
}
