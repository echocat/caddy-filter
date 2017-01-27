package test

import (
	"fmt"
	"github.com/mholt/caddy"
)

// NewTestingCaddy creates a new instance of a caddy with a testing resource for testing purposes.
func NewTestingCaddy(configurationName string) *caddy.Instance {
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
	return instance
}
