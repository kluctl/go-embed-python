package internal

import (
	"os"
	"path/filepath"
	"sync"
)

var createTmpBaseDirOnce sync.Once

func GetTmpBaseDir() string {
	dir := filepath.Join(os.TempDir(), "kluctl-workdir")
	createTmpBaseDirOnce.Do(func() {
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			panic(err)
		}
	})
	return dir
}
