package main

import (
	"github.com/kluctl/go-embed-python/python"
	"os"
)

func main() {
	ep, err := python.NewEmbeddedPython("example")
	if err != nil {
		panic(err)
	}

	cmd := ep.PythonCmd("-c", "print('hello')")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
