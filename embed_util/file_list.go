package embed_util

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

type fileList struct {
	ContentHash string `json:"contentHash"`
	Files       []fileListEntry
}

type fileListEntry struct {
	Name       string      `json:"name"`
	Size       int64       `json:"size"`
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

func buildFileList(dir string) (*fileList, error) {
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
			Name: relPath,
			Size: info.Size(),
			Mode: info.Mode(),
		}

		if info.Mode().Type() == fs.ModeSymlink {
			sl, err := os.Readlink(path)
			if err != nil {
				return err
			}
			fle.Symlink = sl
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
