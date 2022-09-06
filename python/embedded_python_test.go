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
