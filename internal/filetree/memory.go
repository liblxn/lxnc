package filetree

import "strings"

type memFileTree struct {
	files map[string]string // path => contents
}

// Memory returns a memory file system with the given files.
// This file tree type is meant to be used in tests.
func Memory(files map[string]string) FileTree {
	return memFileTree{files: files}
}

func (t memFileTree) Close() error {
	return nil
}

func (t memFileTree) Walk(root string, fn WalkFunc) error {
	for path, contents := range t.files {
		if !strings.HasPrefix(path, root) {
			continue
		}

		err := fn(path, strings.NewReader(contents))
		if err != nil {
			return err
		}
	}
	return nil
}
