package embed

import (
	"fmt"
	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/internal"
	"path/filepath"
	"runtime"
)

var embeddedPythonPath string

func init() {
	embeddedPythonPath = decompressPython()
}

func GetEmbeddedPythonPath() string {
	return embeddedPythonPath
}

func decompressPython() string {
	path := filepath.Join(internal.GetTmpBaseDir(), "python", fmt.Sprintf("python-%s", runtime.GOOS))
	path, err := embed_util.ExtractEmbeddedToTmp(pythonLib, path)
	if err != nil {
		panic(err)
	}

	return path
}
