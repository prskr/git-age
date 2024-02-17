package ports

import (
	"io"
	"io/fs"
	"os"
)

type ReadWriteFile interface {
	fs.File
	io.WriteSeeker
	WriteString(s string) (int, error)
	Name() string
}

type OpenRWOptions func(flags int) int

func WithAppend(flags int) int {
	return flags | os.O_APPEND
}

func WithTruncate(flags int) int {
	return flags | os.O_TRUNC
}

type ReadWriteFS interface {
	fs.FS
	OpenRW(filePath string, opts ...OpenRWOptions) (ReadWriteFile, error)
	Mkdir(dir string, all bool, mode os.FileMode) error
	TempFile(dir, pattern string) (ReadWriteFile, error)
	Rename(oldPath, newPath string) error
	Remove(filePath string) error
}
