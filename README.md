# Embedded Python Interpreter for Go

This library provides an embedded distribution of Python, which should work out-of-the box on a selected set of
architectures and operating systems.

This library does not require CGO and solely relies on executing Python inside another process. It does not rely
on CPython binding to work. There is also no need to have Python pre-installed on the target host.

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

## Supported architectures
The following operating systems and architectures are supported:
* darwin-amd64
* darwin-arm64
* linux-amd64
* linux-arm64
* windows-amd64

## Releases
Releases in this library are handled a bit different from what one might be used to. This library does currently not
follow a versioning schema comparable to sematic versioning. This might however change in the future.

Right now, every tagged release is compromised of the Python interpreter version, the [python-standalone](https://github.com/indygreg/python-build-standalone)
and a build number. For example, the release version `v0.0.0-3.11.6-20231002-2` belongs to Python version 3.11.6, 
the [20231002](https://github.com/indygreg/python-build-standalone/releases/tag/20231002) version of python-standalone
and build number 2. The release version currently always has v0.0.0 as its own version.

The way versioning is handled might result in popular dependency management tools (e.g. dependabot) to not work as you
might require it. Please watch out to not accidentally upgrade your Python version!

## How it works
This library uses the standalone Python distributions found at https://github.com/indygreg/python-build-standalone as
the base.

The `./hack/build-tag.sh` script is used to invoke `python/generate` and `pip/generate`, which then downloads, extracts
and packages all supported Python distributions. The script then also creates a tag which then can be used as a dependency
in your project.

The tagged release internally embed all Python sources and binaries via `//go:embed`. The `EmbeddedPython` object
is then used as a helper utility to access the embedded distribution.

`EmbeddedPython` is created via `NewEmbeddedPython`, which will extract the embedded distribution into a temporary folder.
Extraction is optimized in a way that it is only executed when needed (by verifying integrity of previously extracted
distributions).

## Upgrading python
The Python version and downloaded distributions are controlled via the `.github/workflows/release.yaml` workflow. It
contains a matrix of supported distributions. To upgrade Python, edit this workflow and create a pull request.

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

# Why another go+python solution?
There are already multiple implementations of go-bindings for Python, which however all rely on CGO and/or dynamic
linking. I experimented a lot with these and was not able to make it stable enough so that I could use it without fear
of the process crashing after some time. I even got to the point where I implemented my own dynamic library loader that
was not depending on CGO, but ultimately gave up when I realized that it would not work on all platforms.

The only solution that was left was to spawn a Python process and use some kind of inter-process communication. For this
to work reliably, without any dependencies on the host system, it was required to embed a fully working Python
distribution into my Go binaries. I managed to make this flexible enough to put into a library so that others might
benefit as well.

Initially, this approach/code was part of https://github.com/kluctl/kluctl to allow Jinja2 templates in Go. The Jinja2
part can now be found in https://github.com/kluctl/go-jinja2.
