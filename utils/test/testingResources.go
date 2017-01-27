package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// TestingResourceContentOf load the content of a named testing resource from resources/test/<name>
func TestingResourceContentOf(name string) []byte {
	f := TestingResourceOf(name)
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(fmt.Sprintf("Could not read testing resource '%s'. Got: %v", name, err))
	}
	return bytes
}

// TestingResourceOf load testing resource as stream from resources/test/<name>
func TestingResourceOf(name string) *os.File {
	f, err := os.Open(filepath.Join("resources", "test", name))
	if err != nil {
		panic(fmt.Sprintf("Could not open testing resource '%s'. Got: %v", name, err))
	}
	return f
}
