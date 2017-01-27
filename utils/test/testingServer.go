package test

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// TestingServer represents a http server for testing purposes
type TestingServer struct {
	mux      *http.ServeMux
	server   *http.Server
	listener net.Listener
}

// NewTestingServer creates a new http server for testing purposes
func NewTestingServer(port int) *TestingServer {
	var err error
	result := &TestingServer{}

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

	go func(instance *TestingServer) {
		err := instance.server.Serve(instance.listener)
		if err != nil && !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(fmt.Sprintf("Problem while serving. Got: %v", err))
		}
	}(result)
	return result
}

func (instance *TestingServer) handleDefaultRequest(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	resp.Write([]byte("Hello world!"))
}

// Close closes the testing server graceful.
func (instance *TestingServer) Close() {
	defer func() {
		instance.listener = nil
		instance.server = nil
		instance.mux = nil
	}()
	if instance.listener != nil {
		instance.listener.Close()
	}
}
