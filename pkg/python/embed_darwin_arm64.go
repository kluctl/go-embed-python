package python

import (
	"embed"
	"io/fs"
)

//go:embed all:embed/python-darwin-arm64
var _pythonLib embed.FS
var pythonLib, _ = fs.Sub(_pythonLib, "embed/python-darwin-arm64")
