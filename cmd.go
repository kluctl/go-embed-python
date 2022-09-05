package go_embed_python

import (
	"fmt"
	"github.com/kluctl/go-embed-python/embed"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func PythonCmd(args []string) *exec.Cmd {
	var exePath string
	if runtime.GOOS == "windows" {
		exePath = filepath.Join(embed.GetEmbeddedPythonPath(), "python.exe")
	} else {
		exePath = filepath.Join(embed.GetEmbeddedPythonPath(), "bin/python3")
	}

	cmd := exec.Command(exePath, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONHOME=%s", embed.GetEmbeddedPythonPath()))

	return cmd
}
