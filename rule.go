package filter

import (
	"net/http"
	"regexp"
)

type rule struct {
	path          *regexp.Regexp
	contentType   *regexp.Regexp
	searchPattern *regexp.Regexp
	replacement   []byte
}

func (instance *rule) matches(request *http.Request, responseHeader *http.Header) bool {
	if instance.path != nil && request != nil && instance.path.MatchString(request.URL.Path) {
		return true
	}
	if instance.contentType != nil && responseHeader != nil && instance.contentType.MatchString(responseHeader.Get("Content-Type")) {
		return true
	}
	return false
}

func (instance *rule) execute(request *http.Request, responseHeader *http.Header, input []byte) []byte {
	pattern := instance.searchPattern
	if pattern == nil {
		return input
	}
	action := &ruleReplaceAction{
		request:        request,
		responseHeader: responseHeader,
		searchPattern:  instance.searchPattern,
		replacement:    instance.replacement,
	}
	output := pattern.ReplaceAllFunc(input, action.replacer)
	return output
}
