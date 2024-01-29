package infrastructure

import (
	"github.com/prskr/git-age/core/ports"
	"io/fs"
	"os"
	"path/filepath"
)

var _ ports.ReadWriteFS = (*ReadWriteDirFS)(nil)

func NewReadWriteDirFS(path string) *ReadWriteDirFS {
	return &ReadWriteDirFS{
		path:       path,
		underlying: os.DirFS(path),
	}
}

type ReadWriteDirFS struct {
	path       string
	underlying fs.FS
}

func (f ReadWriteDirFS) Append(filePath string) (ports.ReadWriteFile, error) {
	return os.OpenFile(filepath.Join(f.path, filePath), os.O_APPEND|os.O_WRONLY, 0644)
}

func (f ReadWriteDirFS) Open(name string) (fs.File, error) {
	return f.underlying.Open(name)
}

func (f ReadWriteDirFS) Create(filePath string) (ports.ReadWriteFile, error) {
	return os.Create(filepath.Join(f.path, filePath))
}
