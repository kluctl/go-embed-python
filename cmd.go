package go_embed_python

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func PythonCmd(args []string) *exec.Cmd {
	var exePath string
	if runtime.GOOS == "windows" {
		exePath = filepath.Join(GetEmbeddedPythonPath(), "python.exe")
	} else {
		exePath = filepath.Join(GetEmbeddedPythonPath(), "bin/python3")
	}

	cmd := exec.Command(exePath, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONHOME=%s", GetEmbeddedPythonPath()))

	return cmd
}
