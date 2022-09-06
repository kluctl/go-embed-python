package embed_util

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
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

	b, err := json.MarshalIndent(fl, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(out, "files.json"), b, 0o644)
	return err
}

func copyFiles(out string, dir string, fl *fileList) error {
	hash := sha256.New()
	for _, fle := range fl.Files {
		path := filepath.Join(dir, fle.Name)
		outPath := filepath.Join(out, fle.Name)
		err := os.MkdirAll(filepath.Dir(outPath), 0o755)
		if err != nil {
			return err
		}
		st, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if st.Mode().Type() == os.ModeSymlink {
			sl, err := os.Readlink(path)
			if err != nil {
				return err
			}
			_ = binary.Write(hash, binary.LittleEndian, "symlink")
			_ = binary.Write(hash, binary.LittleEndian, sl)
		} else if st.Mode().IsDir() {
			err = os.MkdirAll(path, fle.Mode.Perm())
			if err != nil {
				return err
			}
			_ = binary.Write(hash, binary.LittleEndian, "dir")
			_ = binary.Write(hash, binary.LittleEndian, fle.Name)
		} else if st.Mode().IsRegular() {
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
			_ = binary.Write(hash, binary.LittleEndian, "regular")
			_ = binary.Write(hash, binary.LittleEndian, fle.Name)
			_ = binary.Write(hash, binary.LittleEndian, data)
		}
	}
	fl.ContentHash = hex.EncodeToString(hash.Sum(nil))
	return nil
}
