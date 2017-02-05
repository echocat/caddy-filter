package test

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"github.com/echocat/caddy-filter/utils/fcgi"
	"log"
)

// TestingFcgiServer represents a http server for testing purposes
type TestingFcgiServer struct {
	mux      *http.ServeMux
	listener net.Listener
}

// NewTestingFcgiServer creates a new http server for testing purposes
func NewTestingFcgiServer(port int) *TestingFcgiServer {
	var err error
	result := &TestingFcgiServer{}

	result.mux = http.NewServeMux()
	result.mux.HandleFunc("/index.cgi", result.handleIndexRequest)
	result.mux.HandleFunc("/redirect.cgi", result.handleRedirectRequest)
	result.mux.HandleFunc("/another.cgi", result.handleAnotherRequest)

	result.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("Could not start test server. Got: %v", err))
	}

	go func(instance *TestingFcgiServer) {
		err := fcgi.Serve(instance.listener, instance.mux)
		if err != nil && !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(fmt.Sprintf("Problem while serving. Got: %v", err))
		}
	}(result)
	return result
}

func (instance *TestingFcgiServer) handleIndexRequest(resp http.ResponseWriter, req *http.Request) {
	fr, ok := resp.(fcgi.ResponseWriter)
	if !ok {
		log.Fatal("ResponseWriter is not the FCGI specific one.")
	}

	resp.WriteHeader(200)
	fr.WriteErr([]byte("Hello from FCGI to server."))
	resp.Write([]byte("<html>" +
		"<head><title>Hello world!</title></head>" +
		"<body><p>Hello world!</p></body>" +
		"</html>"))
}

func (instance *TestingFcgiServer) handleRedirectRequest(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Location", "/another.cgi")
	resp.WriteHeader(301)
	resp.Write([]byte("<a href=\"/another.cgi\">Moved Permanently</a>."))
}

func (instance *TestingFcgiServer) handleAnotherRequest(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("<html>" +
		"<head><title>I'am another!</title></head>" +
		"<body><p>I'am another!</p></body>" +
		"</html>"))
}

// Close closes the testing server graceful.
func (instance *TestingFcgiServer) Close() {
	defer func() {
		instance.listener = nil
		instance.mux = nil
	}()
	if instance.listener != nil {
		instance.listener.Close()
	}
}
