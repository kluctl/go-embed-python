package embed_util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/kluctl/kluctl-python-deps/pkg/utils"
	"github.com/rogpeppe/go-internal/lockedfile"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func ExtractEmbeddedToTmp(embedFs fs.FS, targetPrefix string) (string, error) {
	flStr, err := fs.ReadFile(embedFs, "files.json")
	if err != nil {
		return "", err
	}
	flHash := utils.Sha256Bytes(flStr)

	fl, err := readFileList(string(flStr))
	if err != nil {
		return "", err
	}

	targetPath := fmt.Sprintf("%s-%s", targetPrefix, flHash[:16])

	err = os.MkdirAll(filepath.Dir(targetPath), 0o755)
	if err != nil {
		return "", err
	}

	lock, err := lockedfile.Create(targetPath + ".lock")
	if err != nil {
		return "", err
	}
	defer lock.Close()

	err = os.MkdirAll(targetPath, 0o755)
	if err != nil {
		return "", err
	}

	err = copyEmbeddedFilesToTmp(embedFs, targetPath, fl)
	if err != nil {
		return "", err
	}

	return targetPath, nil
}

func copyEmbeddedFilesToTmp(embedFs fs.FS, targetPath string, fl *fileList) error {
	m := make(map[string]fileListEntry)

	for _, fle := range fl.Files {
		m[fle.Name] = fle
	}

	for _, fle := range fl.Files {
		resolvedFle := fle
		for resolvedFle.Mode.Type() == fs.ModeSymlink {
			if filepath.IsAbs(resolvedFle.Symlink) {
				return fmt.Errorf("abs path not allowed: %s", resolvedFle.Symlink)
			}
			sl := filepath.Clean(filepath.Join(filepath.Dir(resolvedFle.Name), resolvedFle.Symlink))
			fle2, ok := m[sl]
			if !ok {
				return fmt.Errorf("symlink %s at %s could not be resolved", resolvedFle.Symlink, resolvedFle.Name)
			}
			resolvedFle = fle2
			if resolvedFle.Mode.IsDir() {
				return fmt.Errorf("symlinked dirs not supported at the moment: %s -> %s", fle.Name, resolvedFle.Name)
			}
		}

		path := filepath.Join(targetPath, fle.Name)
		existingSt, err := os.Lstat(path)
		if err == nil {
			if resolvedFle.Mode.Type() == existingSt.Mode().Type() {
				if resolvedFle.Mode.IsDir() {
					continue
				} else if existingSt.Size() == resolvedFle.Size {
					// unchanged
					continue
				}
			}
			err = os.RemoveAll(path)
			if err != nil {
				return err
			}
		}

		if fle.Mode.IsDir() {
			err := os.MkdirAll(path, resolvedFle.Mode.Perm())
			if err != nil {
				return err
			}
			continue
		} else if !resolvedFle.Mode.IsRegular() {
			continue
		}

		var data []byte

		if resolvedFle.Compressed {
			data, err = fs.ReadFile(embedFs, resolvedFle.Name+".gz")
			if err != nil {
				return err
			}
			gz, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				return err
			}
			data, err = io.ReadAll(gz)
			_ = gz.Close()
			if err != nil {
				return err
			}
		} else {
			data, err = fs.ReadFile(embedFs, resolvedFle.Name)
			if err != nil {
				return err
			}
		}

		err = os.WriteFile(path, data, resolvedFle.Mode.Perm())
		if err != nil {
			return err
		}
	}

	return nil
}
