package embed_util

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"strings"
)

func CopyForEmbed(out string, dir string) error {
	fl, err := buildFileListFromDir(dir)
	if err != nil {
		return err
	}

	log.Infof("copying to %s with %d files", out, len(fl.Files))
	err = copyFiles(out, dir, fl)
	if err != nil {
		return err
	}

	return doWriteFilesList(out, fl)
}

func BuildAndWriteFilesList(dir string) error {
	fl, err := buildFileListFromDir(dir)
	if err != nil {
		return err
	}
	return doWriteFilesList(dir, fl)
}

func doWriteFilesList(dir string, fl *fileList) error {
	var err error
	fl.ContentHash, err = calcContentHash(dir, fl)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(fl, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "files.json"), b, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func WriteEmbedGoFile(targetDir string, goOs string, goArch string) error {
	var embedSrc, fname string
	if goOs == "" {
		embedSrc = `
package data

import "embed"

//go:embed all:*
var Data embed.FS
`
		fname = "embed.go"
	} else {
		embedSrc = fmt.Sprintf(`
package data

import (
	"embed"
	"io/fs"
)

//go:embed all:%s-%s
var _data embed.FS
var Data, _ = fs.Sub(_data, "%s-%s")
`, goOs, goArch, goOs, goArch)
		fname = strings.ReplaceAll(fmt.Sprintf("embed_%s_%s.go", goOs, goArch), "-", "_")
	}

	return os.WriteFile(filepath.Join(targetDir, fname), []byte(embedSrc), 0o644)
}

func copyFiles(out string, dir string, fl *fileList) error {
	var g errgroup.Group
	g.SetLimit(8)

	for _, fle := range fl.Files {
		fle := fle
		path := filepath.Join(dir, fle.Name)

		st, err := os.Lstat(path)
		if err != nil {
			return err
		}

		outPath := filepath.Join(out, fle.Name)
		err = os.MkdirAll(filepath.Dir(outPath), 0o755)
		if err != nil {
			return err
		}

		if !st.Mode().IsRegular() {
			continue
		}

		g.Go(func() error {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if fle.Compressed {
				b := bytes.NewBuffer(make([]byte, 0, len(data)))
				gz, err := gzip.NewWriterLevel(b, gzip.BestCompression)
				if err != nil {
					return err
				}
				_, err = gz.Write(data)
				if err != nil {
					_ = gz.Close()
					return err
				}
				err = gz.Flush()
				_ = gz.Close()
				if err != nil {
					return err
				}
				data = b.Bytes()
				outPath += ".gz"
			}

			err = os.WriteFile(outPath, data, 0o644)
			if err != nil {
				return err
			}
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return err
	}

	return nil
}

func calcContentHash(dir string, fl *fileList) (string, error) {
	hash := sha256.New()
	for _, fle := range fl.Files {
		path := filepath.Join(dir, fle.Name)
		st, err := os.Lstat(path)
		if err != nil {
			return "", err
		}
		if st.Mode().Type() == os.ModeSymlink {
			sl, err := os.Readlink(path)
			if err != nil {
				return "", err
			}
			_ = binary.Write(hash, binary.LittleEndian, "symlink")
			_ = binary.Write(hash, binary.LittleEndian, sl)
		} else if st.Mode().IsDir() {
			err = os.MkdirAll(path, fle.Mode.Perm())
			if err != nil {
				return "", err
			}
			_ = binary.Write(hash, binary.LittleEndian, "dir")
			_ = binary.Write(hash, binary.LittleEndian, fle.Name)
		} else if st.Mode().IsRegular() {
			outPath := filepath.Join(dir, fle.Name)
			if fle.Compressed {
				outPath += ".gz"
			}

			data, err := os.ReadFile(outPath)
			if err != nil {
				return "", err
			}

			_ = binary.Write(hash, binary.LittleEndian, "regular")
			_ = binary.Write(hash, binary.LittleEndian, fle.Name)
			_ = binary.Write(hash, binary.LittleEndian, data)
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
