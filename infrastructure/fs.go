package infrastructure

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/prskr/git-age/core/ports"
)

var (
	_ ports.ReadWriteFS = (*ReadWriteDirFS)(nil)
	_ fs.ReadDirFS      = (*ReadWriteDirFS)(nil)
)

func NewReadWriteDirFS(rootPath string) *ReadWriteDirFS {
	return &ReadWriteDirFS{
		rootPath:   rootPath,
		underlying: os.DirFS(rootPath),
	}
}

type ReadWriteDirFS struct {
	rootPath   string
	underlying fs.FS
}

func (f ReadWriteDirFS) Mkdir(dir string, all bool, mode os.FileMode) error {
	return os.MkdirAll(filepath.Join(f.rootPath, filepath.FromSlash(dir)), mode)
}

func (f ReadWriteDirFS) Rename(oldPath, newPath string) error {
	fullOldPath := filepath.Join(f.rootPath, filepath.FromSlash(oldPath))
	fullNewPath := filepath.Join(f.rootPath, filepath.FromSlash(newPath))
	return os.Rename(fullOldPath, fullNewPath)
}

func (f ReadWriteDirFS) Remove(filePath string) error {
	return os.Remove(filepath.Join(f.rootPath, filepath.FromSlash(filePath)))
}

func (f ReadWriteDirFS) TempFile(dir, pattern string) (ports.ReadWriteFile, error) {
	tmpFile, err := os.CreateTemp(filepath.Join(f.rootPath, filepath.FromSlash(dir)), pattern)
	if err != nil {
		return nil, err
	}

	return readWriteOsFile{File: tmpFile, fsRoot: f.rootPath}, nil
}

func (f ReadWriteDirFS) Create(filePath string, opts ...ports.OpenRWOptions) (ports.ReadWriteFile, error) {
	openFlags := os.O_CREATE | os.O_RDWR

	for _, opt := range opts {
		openFlags = opt(openFlags)
	}

	file, err := os.OpenFile(filepath.Join(f.rootPath, filepath.FromSlash(filePath)), openFlags, 0o644)
	if err != nil {
		return nil, err
	}

	return readWriteOsFile{File: file, fsRoot: f.rootPath}, nil
}

func (f ReadWriteDirFS) Open(name string) (fs.File, error) {
	return f.underlying.Open(name)
}

func (f ReadWriteDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(filepath.Join(f.rootPath, filepath.FromSlash(name)))
}

var _ ports.ReadWriteFile = (*readWriteOsFile)(nil)

type readWriteOsFile struct {
	fsRoot string
	*os.File
}

func (r readWriteOsFile) Name() string {
	name := r.File.Name()

	var err error
	name, err = filepath.Rel(r.fsRoot, name)
	if err != nil {
		panic("failed to get relative path: " + err.Error())
	}

	return filepath.ToSlash(name)
}
