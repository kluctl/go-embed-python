package main

import (
	"os"

	"github.com/kluctl/go-embed-python/python"
)

func main() {
	ep, err := python.NewEmbeddedPython("example")
	if err != nil {
		panic(err)
	}

	cmd, err := ep.PythonCmd("-c", "print('hello')")
	if err != nil {
		panic(err)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
