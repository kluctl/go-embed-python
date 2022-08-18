package python

import "embed"

//go:embed all:embed/python-linux-amd64
var _pythonLib embed.FS
var pythonLib, _ = fs.Sub(_pythonLib, "embed/python-linux-amd64")
