package python

import (
	"fmt"
	"github.com/kluctl/kluctl-python-deps/pkg/embed_util"
	"github.com/kluctl/kluctl-python-deps/pkg/utils"
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
	path := filepath.Join(utils.GetTmpBaseDir(), "python", fmt.Sprintf("python-%s", runtime.GOOS))
	path, err := embed_util.ExtractEmbeddedToTmp(pythonLib, path)
	if err != nil {
		panic(err)
	}

	return path
}
