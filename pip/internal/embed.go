package internal

import (
	"embed"
	"io/fs"
)

//go:generate go run ./generate

//go:embed all:data/pip
var _pipLib embed.FS
var PipLib, _ = fs.Sub(_pipLib, "data/pip")
