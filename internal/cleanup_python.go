package internal

import (
	"github.com/gobwas/glob"
	"io/fs"
	"os"
	"path/filepath"
)

var DefaultPythonRemovePatterns = []glob.Glob{
	glob.MustCompile("__pycache__"),
	glob.MustCompile("**/__pycache__"),
	glob.MustCompile("**.a"),
	glob.MustCompile("**.pdb"),
	glob.MustCompile("**.pyc"),
	glob.MustCompile("**/test_*.py"),
	glob.MustCompile("**/*.dist-info"),
}

func CleanupPythonDir(dir string, keepPatterns []glob.Glob) error {
	var removes []string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		for _, p := range DefaultPythonRemovePatterns {
			if p.Match(relPath) {
				removes = append(removes, path)
			}
		}
		if len(keepPatterns) != 0 && !info.Mode().IsDir() {
			keep := false
			for _, p := range keepPatterns {
				if p.Match(relPath) {
					keep = true
					break
				}
			}
			if !keep {
				removes = append(removes, path)
			}
		}
		return nil
	})

	for _, r := range removes {
		err = os.RemoveAll(r)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	err = removeEmptyDirs(dir)
	if err != nil {
		return err
	}

	return err
}

func removeEmptyDirs(dir string) error {
	for true {
		didRemove, err := removeEmptyDirs2(dir)
		if err != nil {
			return err
		}
		if !didRemove {
			break
		}
	}
	return nil
}

func removeEmptyDirs2(dir string) (bool, error) {
	var removes []string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			des, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			if len(des) == 0 {
				removes = append(removes, path)
			}
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	for _, r := range removes {
		err = os.Remove(r)
		if err != nil {
			return false, err
		}
	}
	return len(removes) != 0, nil
}
