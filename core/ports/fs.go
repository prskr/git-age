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

type ReadWriteFS interface {
	fs.FS
	OpenRW(filePath string) (ReadWriteFile, error)
	Mkdir(dir string, all bool, mode os.FileMode) error
	TempFile(dir, pattern string) (ReadWriteFile, error)
	Rename(oldPath, newPath string) error
	Remove(filePath string) error
}
