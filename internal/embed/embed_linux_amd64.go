package embed

import (
	"embed"
	"io/fs"
)

//go:embed all:data/python-linux-amd64
var _pythonLib embed.FS
var PythonLib, _ = fs.Sub(_pythonLib, "data/python-linux-amd64")
