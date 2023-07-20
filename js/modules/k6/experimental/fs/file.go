package fs

import (
	"path/filepath"
)

// file is an abstraction for interacting with files.
type file struct {
	// name holds the name of the file, as presented to [Open].
	path string

	// data holds a pointer to the file's data
	data []byte

	// offset holds the current offset in the file
	offset int
}

// Stat returns a FileInfo describing the named file.
func (f *file) Stat() *FileInfo {
	filename := filepath.Base(f.path)
	return &FileInfo{Name: filename, Size: len(f.data)}
}

// FileInfo holds information about a file.
//
// It is a wrapper around the [fileInfo] struct, which is meant to be directly
// exposed to the JS runtime.
type FileInfo struct {
	// Name holds the base name of the file.
	Name string `json:"name"`

	// Name holds the length in bytes of the file.
	Size int `json:"size"`
}

func (f *file) Read(into []byte) (int, error) {
	start := f.offset
	if start == len(f.data) {
		return 0, newFsError(EOFError, "EOF")
	}

	end := f.offset + len(into)
	if end > len(f.data) {
		end = len(f.data)
	}

	n := copy(into, f.data[start:end])

	f.offset += n

	return n, nil
}
