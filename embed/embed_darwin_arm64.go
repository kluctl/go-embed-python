package embed

import (
	"embed"
	"io/fs"
)

//go:embed all:./data/python-darwin-arm64
var _pythonLib embed.FS
var pythonLib, _ = fs.Sub(_pythonLib, "data/python-darwin-arm64")
