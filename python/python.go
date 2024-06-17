package python

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Python struct {
	pythonHome string
	pythonPath []string
}

func NewPython(pythonHome string) *Python {
	return &Python{
		pythonHome: pythonHome,
	}
}

func (ep *Python) GetBinPath() string {
	if runtime.GOOS == "windows" {
		return ep.pythonHome
	} else {
		return filepath.Join(ep.pythonHome, "bin")
	}
}

func (ep *Python) GetExePath() string {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	} else {
		suffix = "3"
	}
	return filepath.Join(ep.GetBinPath(), "python"+suffix)
}

func (ep *Python) AddPythonPath(p string) {
	ep.pythonPath = append(ep.pythonPath, p)
}

func (ep *Python) PythonCmd(args ...string) *exec.Cmd {
	return ep.PythonCmd2(args)
}

func (ep *Python) PythonCmd2(args []string) *exec.Cmd {
	exePath := ep.GetExePath()

	cmd := exec.Command(exePath, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONHOME=%s", ep.pythonHome))

	if len(ep.pythonPath) != 0 {
		pythonPathEnv := fmt.Sprintf("PYTHONPATH=%s", strings.Join(ep.pythonPath, string(os.PathListSeparator)))
		cmd.Env = append(cmd.Env, pythonPathEnv)
	}

	return cmd
}
