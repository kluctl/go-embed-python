package internal

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractTarStream(r io.Reader, targetPath string) error {
	tarReader := tar.NewReader(r)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("ExtractTarStream: Next() failed: %w", err)
		}

		if !validRelPath(header.Name) {
			return fmt.Errorf("tar contained invalid name error %q", header.Name)
		}

		p := filepath.FromSlash(header.Name)
		p = filepath.Join(targetPath, p)

		err = os.MkdirAll(filepath.Dir(p), 0755)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(p, 0755); err != nil {
				return fmt.Errorf("ExtractTarStream: Mkdir() failed: %w", err)
			}
		case tar.TypeReg:
			_ = os.Remove(p) // we allow overwriting, which easily happens on case insensitive filesystems
			outFile, err := os.Create(p)
			if err != nil {
				return fmt.Errorf("ExtractTarStream: Create() failed: %w", err)
			}
			_, err = io.Copy(outFile, tarReader)
			_ = outFile.Close()
			if err != nil {
				return fmt.Errorf("ExtractTarStream: Copy() failed: %w", err)
			}
			err = os.Chmod(p, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("ExtractTarStream: Chmod() failed: %w", err)
			}
			err = os.Chtimes(p, header.AccessTime, header.ModTime)
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			_ = os.Remove(p) // we allow overwriting, which easily happens on case insensitive filesystems
			if err := os.Symlink(header.Linkname, p); err != nil {
				return fmt.Errorf("ExtractTarStream: Symlink() failed: %w", err)
			}
		default:
			return fmt.Errorf("ExtractTarStream: uknown type %v in %v", header.Typeflag, header.Name)
		}
	}
	return nil
}

func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}
