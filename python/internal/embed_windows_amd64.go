package internal

import (
	"embed"
	"io/fs"
)

//go:embed all:data/python-windows-amd64
var _pythonLib embed.FS
var PythonLib, _ = fs.Sub(_pythonLib, "data/python-windows-amd64")
