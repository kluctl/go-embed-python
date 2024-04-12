package python

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tryiris-ai/go-embed-python/embed_util"
	"github.com/tryiris-ai/go-embed-python/python/internal/data"
)

type EmbeddedPython struct {
	e *embed_util.EmbeddedFiles

	pythonPath []string
}

// NewEmbeddedPython creates a new EmbeddedPython instance. The embedded source code and python binaries are
// extracted on demand using the given name as the base for the temporary directory. You should ensure that the chosen
// name does collide with other consumers of this library.
func NewEmbeddedPython(name string) (*EmbeddedPython, error) {
	e, err := embed_util.NewEmbeddedFiles(data.Data, fmt.Sprintf("python-%s", name))
	if err != nil {
		return nil, err
	}
	return &EmbeddedPython{
		e: e,
	}, nil
}

func NewEmbeddedPythonWithTmpDir(tmpDir string, withHashInDir bool) (*EmbeddedPython, error) {
	e, err := embed_util.NewEmbeddedFilesWithTmpDir(data.Data, tmpDir, withHashInDir)
	if err != nil {
		return nil, err
	}
	return &EmbeddedPython{
		e: e,
	}, nil
}

func (ep *EmbeddedPython) Cleanup() error {
	return ep.e.Cleanup()
}

func (ep *EmbeddedPython) GetExtractedPath() string {
	return ep.e.GetExtractedPath()
}

func (ep *EmbeddedPython) GetBinPath() string {
	if runtime.GOOS == "windows" {
		return ep.GetExtractedPath()
	} else {
		return filepath.Join(ep.GetExtractedPath(), "bin")
	}
}

func (ep *EmbeddedPython) GetExePath() string {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	} else {
		suffix = "3"
	}
	return filepath.Join(ep.GetBinPath(), "python"+suffix)
}

func (ep *EmbeddedPython) AddPythonPath(p string) {
	ep.pythonPath = append(ep.pythonPath, p)
}

func (ep *EmbeddedPython) PythonCmd(args ...string) *exec.Cmd {
	return ep.PythonCmd2(args)
}

func (ep *EmbeddedPython) PythonCmd2(args []string) *exec.Cmd {
	exePath := ep.GetExePath()

	cmd := exec.Command(exePath, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONHOME=%s", ep.GetExtractedPath()))

	if len(ep.pythonPath) != 0 {
		pythonPathEnv := fmt.Sprintf("PYTHONPATH=%s", strings.Join(ep.pythonPath, string(os.PathListSeparator)))
		cmd.Env = append(cmd.Env, pythonPathEnv)
	}

	return cmd
}
