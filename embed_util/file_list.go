package embed_util

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

type fileList struct {
	ContentHash string          `json:"contentHash"`
	Files       []fileListEntry `json:"files"`
}

type fileListEntry struct {
	Name       string      `json:"name"`
	Size       int64       `json:"size"`
	ModTime    int64       `json:"modTime"`
	Mode       fs.FileMode `json:"perm"`
	Symlink    string      `json:"symlink,omitempty"`
	Compressed bool        `json:"compressed,omitempty"`
}

func readFileList(fileListStr string) (*fileList, error) {
	var fl fileList
	err := json.Unmarshal([]byte(fileListStr), &fl)
	if err != nil {
		return nil, err
	}
	return &fl, err
}

func buildFileListFromDir(dir string) (*fileList, error) {
	var fl fileList

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		fle := fileListEntry{
			Name:    relPath,
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode(),
		}

		if info.Mode().Type() == fs.ModeSymlink {
			sl, err := os.Readlink(path)
			if err != nil {
				return err
			}
			fle.Symlink = sl
			fle.Mode &= ^fs.ModePerm
		} else if info.Mode().IsDir() {
			fle.Size = 0
		} else if info.Mode().IsRegular() {
			fle.Compressed = shouldCompress(path)
		}

		fl.Files = append(fl.Files, fle)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(fl.Files, func(i, j int) bool {
		return fl.Files[i].Name < fl.Files[j].Name
	})
	return &fl, nil
}

func buildFileListFromFs(embedFs fs.FS) (*fileList, error) {
	var fl fileList

	err := fs.WalkDir(embedFs, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		fle := fileListEntry{
			Name:    path,
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode() | 0o600,
		}

		if info.Mode().Type() == fs.ModeSymlink {
			return fmt.Errorf("symlink not supported in buildFileListFromFs")
		} else if info.Mode().IsDir() {
			fle.Size = 0
			fle.ModTime = 0
		}

		fl.Files = append(fl.Files, fle)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(fl.Files, func(i, j int) bool {
		return fl.Files[i].Name < fl.Files[j].Name
	})
	return &fl, nil
}

func shouldCompress(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	data := make([]byte, 512)
	n, err := f.Read(data)
	if err != nil {
		return false
	}
	if http.DetectContentType(data[:n]) == "application/octet-stream" {
		return true
	}
	return false
}

func (fl *fileList) toMap() map[string]fileListEntry {
	m := make(map[string]fileListEntry)
	for _, e := range fl.Files {
		m[e.Name] = e
	}
	return m
}

func (fl *fileList) Hash() string {
	h := sha256.New()
	e := json.NewEncoder(h)
	err := e.Encode(fl)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}
