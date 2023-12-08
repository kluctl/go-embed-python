package python

import (
	"bytes"
	"fmt"
	"github.com/kluctl/go-embed-python/internal"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"testing"
)

func TestEmbeddedPython(t *testing.T) {
	rndName := fmt.Sprintf("test-%d", rand.Uint32())
	ep, err := NewEmbeddedPython(rndName)
	assert.NoError(t, err)
	defer ep.Cleanup()
	path := ep.GetExtractedPath()
	assert.NotEqual(t, path, "")
	pexe := ep.GetExePath()
	assert.True(t, internal.Exists(pexe))

	cmd := ep.PythonCmd("-c", "print('test test')")
	stdout, err := cmd.StdoutPipe()
	assert.NoError(t, err)
	defer stdout.Close()

	err = cmd.Start()
	assert.NoError(t, err)

	stdoutStr, err := io.ReadAll(stdout)
	assert.NoError(t, err)

	err = cmd.Wait()
	assert.NoError(t, err)

	stdoutStr = bytes.TrimSpace(stdoutStr)
	assert.Equal(t, "test test", string(stdoutStr))
}

func TestPrintSystemInfo(t *testing.T) {
	getSystemInfo := `
import platform, sys

print("system info:")
print("sys.version=" + sys.version)

print("platform.python_version=" + platform.python_version())
print("platform.machine=" + platform.machine())
print("platform.version=" + platform.version())
print("platform.platform=" + platform.platform())
print("platform.release=" + platform.release())
print("platform.uname=" + str(platform.uname()))
print("platform.system=" + platform.system())
print("platform.processor=" + platform.processor())
`

	rndName := fmt.Sprintf("test-%d", rand.Uint32())
	ep, err := NewEmbeddedPython(rndName)
	assert.NoError(t, err)
	defer ep.Cleanup()
	path := ep.GetExtractedPath()
	assert.NotEqual(t, path, "")
	pexe := ep.GetExePath()
	assert.True(t, internal.Exists(pexe))

	cmd := ep.PythonCmd("-c", getSystemInfo)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	defer stdout.Close()
	defer stderr.Close()

	err = cmd.Start()
	assert.NoError(t, err)

	stdoutStr, _ := io.ReadAll(stdout)
	stderrStr, _ := io.ReadAll(stderr)
	t.Log("stdout=" + string(stdoutStr))
	t.Log("stderr=" + string(stderrStr))

	err = cmd.Wait()
	assert.NoError(t, err)
}
