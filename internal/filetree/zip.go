package filetree

import (
	"archive/zip"
	"path/filepath"
)

// zipFileTree represents a file tree in a zip archive file.
type zipFileTree struct {
	r *zip.ReadCloser
}

func Zip(filename string) (FileTree, error) {
	r, err := zip.OpenReader(filename)
	return zipFileTree{r: r}, err
}

func (t zipFileTree) Close() error {
	return t.r.Close()
}

func (t zipFileTree) Walk(root string, fn WalkFunc) error {
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
