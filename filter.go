package filter

import (
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"net/http"
)

type filterHandler struct {
	Next  httpserver.Handler
	Rules []*rule
}

func (instance filterHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) (int, error) {
	wrapper := newResponseWriterWrapperFor(writer, func(wrapper *responseWriterWrapper) bool {
		for _, rule := range instance.Rules {
			if rule.matches(request, wrapper.Header()) {
				return true
			}
		}
		return false
	})
	result, err := instance.Next.ServeHTTP(wrapper, request)
	if err != nil {
		return result, err
	}
	if !wrapper.wasSomethingRecorded() || !wrapper.isBodyAllowed() {
		return result, nil
	}
	body := wrapper.recorded()
	for _, rule := range instance.Rules {
		if rule.matches(request, wrapper.Header()) {
			body, err = rule.execute(request, wrapper.Header(), body)
			if err != nil {
				return result, err
			}
		}
	}
	_, err = writer.Write(body)
	if err != nil {
		return result, err
	}
	return result, nil
}
