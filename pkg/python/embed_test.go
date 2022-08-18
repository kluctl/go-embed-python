package python

import (
	"github.com/kluctl/kluctl-python-deps/pkg/utils"
	"github.com/stretchr/testify/assert"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEmbeddedPython(t *testing.T) {
	path := GetEmbeddedPythonPath()
	assert.NotEqual(t, path, "")
	pexe := filepath.Join(path, "bin/python3")
	if runtime.GOOS == "windows" {
		pexe += ".exe"
	}
	assert.True(t, utils.Exists(pexe))

	cmd := exec.Command(pexe, "-c", "print('test')")
	stdout, err := cmd.StdoutPipe()
	assert.NoError(t, err)
	defer stdout.Close()

	err = cmd.Start()
	assert.NoError(t, err)

	stdoutStr, err := io.ReadAll(stdout)
	assert.NoError(t, err)

	err = cmd.Wait()
	assert.NoError(t, err)

	assert.Equal(t, "test\n", string(stdoutStr))
}
