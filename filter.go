package filter

import (
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"io"
	"net/http"
	"strconv"
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
	atLeastOneRuleMatched := false
	for _, rule := range instance.rules {
		if rule.matches(request, &header) {
			body = rule.execute(request, &header, body)
			atLeastOneRuleMatched = true
		}
	}
	if atLeastOneRuleMatched {
		oldContentLength := wrapper.Header().Get("Content-Length")
		if len(oldContentLength) > 0 {
			newContentLength := strconv.Itoa(len(body))
			wrapper.Header().Set("Content-Length", newContentLength)
		}
	}
	n, err := wrapper.writeToDelegate(body)
	if err != nil {
		return result, err
	}
	if n < len(body) {
		return result, io.ErrShortWrite
	}
	return result, nil
}
