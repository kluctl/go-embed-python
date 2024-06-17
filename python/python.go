package python

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Python interface {
	GetExeName() string
	GetExePath() (string, error)
	AddPythonPath(p string)
	PythonCmd(args ...string) (*exec.Cmd, error)
	PythonCmd2(args []string) (*exec.Cmd, error)
}

type python struct {
	pythonHome string
	pythonPath []string
}

type PythonOpt func(o *python)

func WithPythonHome(home string) PythonOpt {
	return func(o *python) {
		o.pythonHome = home
	}
}

func NewPython(opts ...PythonOpt) Python {
	ep := &python{}

	for _, o := range opts {
		o(ep)
	}

	return ep
}

func (ep *python) GetExeName() string {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	} else {
		suffix = "3"
	}
	return "python" + suffix
}

func (ep *python) GetExePath() (string, error) {
	if ep.pythonHome == "" {
		p, err := exec.LookPath(ep.GetExeName())
		if err != nil {
			return "", fmt.Errorf("failed to determine %s path: %w", ep.GetExeName(), err)
		}
		return p, nil
	} else {
		var p string
		if runtime.GOOS == "windows" {
			p = filepath.Join(ep.pythonHome, ep.GetExeName())
		} else {
			p = filepath.Join(ep.pythonHome, "bin", ep.GetExeName())
		}
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("failed to determine %s path: %w", ep.GetExeName(), err)
		}
		return p, nil
	}
}

func (ep *python) AddPythonPath(p string) {
	ep.pythonPath = append(ep.pythonPath, p)
}

func (ep *python) PythonCmd(args ...string) (*exec.Cmd, error) {
	return ep.PythonCmd2(args)
}

func (ep *python) PythonCmd2(args []string) (*exec.Cmd, error) {
	exePath, err := ep.GetExePath()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(exePath, args...)
	cmd.Env = os.Environ()

	if ep.pythonHome != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONHOME=%s", ep.pythonHome))
	}

	if len(ep.pythonPath) != 0 {
		pythonPathEnv := fmt.Sprintf("PYTHONPATH=%s", strings.Join(ep.pythonPath, string(os.PathListSeparator)))
		cmd.Env = append(cmd.Env, pythonPathEnv)
	}

	return cmd, nil
}
