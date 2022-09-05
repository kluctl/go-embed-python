package internal

import (
	"archive/tar"
	"fmt"
	securejoin "github.com/cyphar/filepath-securejoin"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

		header.Name = filepath.FromSlash(header.Name)

		p, err := securejoin.SecureJoin(targetPath, header.Name)
		if err != nil {
			return err
		}
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
			if err := os.Symlink(header.Linkname, p); err != nil {
				return fmt.Errorf("ExtractTarStream: Symlink() failed: %w", err)
			}
		default:
			return fmt.Errorf("ExtractTarStream: uknown type %v in %v", header.Typeflag, header.Name)
		}
	}
	return nil
}

func AddToTar(tw *tar.Writer, pth string, name string, filter func(h *tar.Header, size int64) (*tar.Header, error)) error {
	fi, err := os.Lstat(pth)
	if err != nil {
		return err
	}

	var linkName string
	if fi.Mode().Type() == fs.ModeSymlink {
		x, err := os.Readlink(pth)
		if err != nil {
			return err
		}
		linkName = x
	}

	h, err := tar.FileInfoHeader(fi, linkName)
	if err != nil {
		return err
	}
	h.Name = filepath.ToSlash(name)

	if filter != nil {
		s := fi.Size()
		if fi.IsDir() {
			s = 0
		}
		h, err = filter(h, s)
		if err != nil {
			return err
		}
		if h == nil {
			return nil
		}
	}

	err = tw.WriteHeader(h)
	if err != nil {
		return err
	}

	if fi.Mode().Type() == fs.ModeSymlink {
		return nil
	}

	if fi.Mode().IsDir() {
		des, err := os.ReadDir(pth)
		if err != nil {
			return err
		}
		for _, d := range des {
			err = AddToTar(tw, filepath.Join(pth, d.Name()), filepath.Join(name, d.Name()), filter)
			if err != nil {
				return err
			}
		}
		return nil
	} else if fi.Mode().IsRegular() {
		f, err := os.Open(pth)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("unsupported file type/mode %s", fi.Mode().String())
	}
}
