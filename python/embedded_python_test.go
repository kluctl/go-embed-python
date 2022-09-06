package python

import (
	"fmt"
	"github.com/kluctl/go-embed-python/internal"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEmbeddedPython(t *testing.T) {
	rndName := fmt.Sprintf("test-%d", rand.Uint32())
	ep, err := NewEmbeddedPython(rndName)
	assert.NoError(t, err)
	defer ep.Cleanup()
	path := ep.GetExtractedPath()
	assert.NotEqual(t, path, "")
	pexe := filepath.Join(path, "bin/python3")
	if runtime.GOOS == "windows" {
		pexe += ".exe"
	}
	assert.True(t, internal.Exists(pexe))

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
