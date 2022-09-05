package internal

import (
	"crypto/sha256"
	"encoding/hex"
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

func Sha256Bytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
