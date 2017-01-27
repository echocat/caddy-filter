package test

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// TestingHttpServer represents a http server for testing purposes
type TestingHttpServer struct {
	mux      *http.ServeMux
	server   *http.Server
	listener net.Listener
}

// NewTestingHttpServer creates a new http server for testing purposes
func NewTestingHttpServer(port int) *TestingHttpServer {
	var err error
	result := &TestingHttpServer{}

	result.mux = http.NewServeMux()
	result.mux.HandleFunc("/default", result.handleDefaultRequest)

	result.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: result.mux,
	}

	result.listener, err = net.Listen("tcp", result.server.Addr)
	if err != nil {
		panic(fmt.Sprintf("Could not start test server. Got: %v", err))
	}

	go func(instance *TestingHttpServer) {
		err := instance.server.Serve(instance.listener)
		if err != nil && !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(fmt.Sprintf("Problem while serving. Got: %v", err))
		}
	}(result)
	return result
}

func (instance *TestingHttpServer) handleDefaultRequest(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	resp.Write([]byte("Hello world!"))
}

// Close closes the testing server graceful.
func (instance *TestingHttpServer) Close() {
	defer func() {
		instance.listener = nil
		instance.server = nil
		instance.mux = nil
	}()
	if instance.listener != nil {
		instance.listener.Close()
	}
}
