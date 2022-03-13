package filetree

import "io"

type WalkFunc func(path string, r io.Reader) error

type FileTree interface {
	Close() error
	Walk(root string, f WalkFunc) error
}
