# Embedded Python Interpreter for Go

This library provides an embedded distribution of Python, which should work out-of-the box on a selected set of
architectures and operating systems.

This library does not require CGO and solely relies on executing Python inside another process. It does not rely
on CPython binding to work. There is also no need to have Pyton pre-installed on the target host.

You really only have to depend on this library and invoke it as follows:

```go
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
```

## How it works
This library uses the standalone Python distributions found at https://github.com/indygreg/python-build-standalone as
the base.

Running `go generate ./...` from the root of the project will will trigger download, extraction and cleanup of these
distributions for all supported platforms. Cleanup means that everything is removed from the distributions that is not
required to run the Python interpreter.

`//go:embed` is then used to embed the extracted distributions into the compiled go binaries. The `EmbeddedPython` object
is then used as a helper utility to access the embedded distribution.

`EmbeddedPython` is created via `NewEmbeddedPython`, which will extract the embedded distribution into a temporary folder.
Extraction is optimized in a way that it is only executed when needed (by verifying integrity of previously extracted
distributions).

## Upgrading python
The Python version and downloaded distributions are controlled via `python/internal/generate/main.go`. To upgrade
Python, increase the version number and re-run `go generate ./...`. Commit the result to git and then create a release.

## Embedding Python libraries into your applications
This library provides utilities/helpers to allow embedding of external libraries into your own application.

To do this, create a simple generator application inside your application/library, for example in `internal/my-python-libs/generate/main.go`:

```go
package main

import (
	"github.com/kluctl/go-embed-python/pip"
)

func main() {
	err := pip.CreateEmbeddedPipPackagesForKnownPlatforms("requirements.txt", "./data/")
	if err != nil {
		panic(err)
	}
}
```

Then create add the `//go:generate go run ./generate` statement to a .go file above the generator source, e.g. in `internal/my-python-libs/dummy.go`:
```
package internal

//go:generate go run ./generate
```

And the requirements.txt in `internal/my-python-libs/requirements.txt`:
```
jinja2==3.1.2
```

When running `go generate ./...` inside your application/library, you'll get the referenced Python libraries installed
to `internal/my-python-libs/data`. The embedded data is then available via `data.Data` and can be passed to
`embed_util.NewEmbeddedFiles()` for extraction.

The path returned by `EmbeddedFiles.GetExtractedPath()` can then be added to the `EmbeddedPython` by calling
`AddPythonPath` on it.

An example of all this can be found in https://github.com/kluctl/go-jinja2
