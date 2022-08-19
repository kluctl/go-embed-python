# kluctl-python-deps

This repository holds an embedded distribution of Python and dependencies required to run the Jinja2 interpreter embedded
into kluctl.

## How it works
The application under pkg/generate will download Python from https://github.com/indygreg/python-build-standalone, extract
it and then clean up unneeded stuff, e.g. libraries that are not used, test resources, documentation, and so on. The
result is copied to pkg/python/embed and then embedded into go via the pkg/python/embed_*.go files.

`GetEmbeddedPythonPath()` can then be used by kluctl to get the path the unpacked python distribution. Unpacking is
performed on startup and is optimized to only write files to TMP when really needed (they don't exist).

## Upgrading python
To upgrade/re-embed Python, run `go generate ./...`.

## Upgrading Jinja2 dependencies
pkg/generate will also run `pip install -r requirements.txt -t <path-to-libs>` on every Python distribution for all
platforms.

You can manually edit requirements.txt to update Jinja2 dependencies. Re-run `go generate ./...` afterwards.