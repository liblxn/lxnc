package filetree

import (
	"os"
	"path/filepath"
)

type osFileTree struct {
	path string
}

// New returns a file tree under a plain directory. The specified path
// is the path to the directory which should be walked.
func New(path string) FileTree {
	return osFileTree{path: path}
}

func (t osFileTree) Close() error {
	return nil
}

func (t osFileTree) Walk(root string, fn WalkFunc) error {
	root = filepath.Join(t.path, root)
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
