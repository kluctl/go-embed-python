package internal

import (
	"embed"
)

//go:generate go run ./generate

//go:embed all:data/pip
var PipLib embed.FS
