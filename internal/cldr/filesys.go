package cldr

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

type walkFunc func(path string, r io.Reader) error

type fileTree interface {
	Close() error
	Walk(root string, f walkFunc) error
}

// osFileTree represents a file tree under a plain directory. The
// string specifies the directory path.
type osFileTree string

func newOsFileTree(directory string) osFileTree {
	return osFileTree(directory)
}

func (t osFileTree) Close() error {
	return nil
}

func (t osFileTree) Walk(root string, fn walkFunc) error {
	root = filepath.Join(string(t), root)
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(info.Name()) != ".xml" {
			return err
		}

		relpath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return fn(relpath, f)
	})
}

// zipFileTree represents a file tree in a zip archive file.
type zipFileTree struct {
	r *zip.ReadCloser
}

func newZipFileTree(filename string) (t zipFileTree, err error) {
	t.r, err = zip.OpenReader(filename)
	return
}

func (t zipFileTree) Close() error {
	return t.r.Close()
}

func (t zipFileTree) Walk(root string, fn walkFunc) error {
	for _, f := range t.r.File {
		path, err := filepath.Rel(root, f.Name)
		if err != nil {
			continue // ignore file not under root
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		err = fn(path, rc)
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
