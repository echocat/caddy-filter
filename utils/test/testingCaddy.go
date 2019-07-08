package test

import (
	"flag"
	"fmt"
	"github.com/caddyserver/caddy"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

var (
	caddyBinary string
)

func init() {
	flag.StringVar(&caddyBinary, "integrationTest.caddyBinary", "", "Specifies the caddy binary to use for integration tests instead of embedded version.")
	flag.Parse()
	if len(caddyBinary) > 0 {
		log.Printf("Using caddy binary '%s' for integration tests instead of embedded version.", caddyBinary)
	}
}

// NewTestingCaddy creates a new instance of a caddy with a testing resource for testing purposes.
func NewTestingCaddy(configurationName string) io.Closer {
	if len(caddyBinary) <= 0 {
		return newEmbeddedTestingCaddy(configurationName)
	}
	return newExternalTestingCaddy(caddyBinary, configurationName)
}

type embeddedTestingCaddy struct {
	caddy *caddy.Instance
}

func newEmbeddedTestingCaddy(configurationName string) *embeddedTestingCaddy {
	config := TestingResourceOf(configurationName)
	defer config.Close()
	caddyFile, err := caddy.CaddyfileFromPipe(config, "http")
	if err != nil {
		panic(fmt.Sprintf("Could not read config '%s'. Got: %v", configurationName, err))
	}
	instance, err := caddy.Start(caddyFile)
	if err != nil {
		panic(fmt.Sprintf("Could not start caddy with config '%s'. Got: %v", configurationName, err))
	}
	return &embeddedTestingCaddy{
		caddy: instance,
	}
}

func (instance *embeddedTestingCaddy) Close() error {
	return instance.caddy.Stop()
}

type externalTestingCaddy struct {
	cmd *exec.Cmd
}

func newExternalTestingCaddy(caddyBinary string, configurationName string) *externalTestingCaddy {
	cmd := exec.Command(caddyBinary, "-agree", "-conf", TestingPathOfResource(configurationName))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		panic(fmt.Sprintf("Could not start caddy process. Got: %v", err))
	}
	time.Sleep(200 * time.Millisecond)
	return &externalTestingCaddy{
		cmd: cmd,
	}
}

func (instance *externalTestingCaddy) Close() error {
	return instance.cmd.Process.Kill()
}
