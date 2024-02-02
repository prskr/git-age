package testfs

import (
	"bytes"
	"io/fs"
	"math/rand"
	"strings"

	"github.com/prskr/git-age/core/ports"
)

var _ ports.ReadWriteFS = (*TestFS)(nil)

func NewTestFS() TestFS {
	return TestFS{
		content: make(map[string]any),
	}
}

type TestFS struct {
	relativePath string
	content      map[string]any
}

func (t TestFS) Open(name string) (fs.File, error) {
	return t.open(strings.Split(name, "/"))
}

func (t TestFS) OpenRW(filePath string) (ports.ReadWriteFile, error) {
	return t.open(strings.Split(filePath, "/"))
}

func (t TestFS) TempFile(dir, pattern string) (ports.ReadWriteFile, error) {
	suffix := randomString(8)

	var fileName string
	if strings.ContainsRune(pattern, '*') {
		fileName = strings.Replace(pattern, "*", suffix, 1)
	} else {
		fileName = pattern + suffix
	}

	tmpFileSegments := append(strings.Split(dir, "/"), fileName)

	if err := t.add(nil, tmpFileSegments, nil); err != nil {
		return nil, err
	}

	return t.open(tmpFileSegments)
}

func (t TestFS) Rename(oldPath, newPath string) error {
	f, err := t.open(strings.Split(oldPath, "/"))
	if err != nil {
		return err
	}

	return t.add(nil, strings.Split(newPath, "/"), f.(*testFile).content.Bytes())
}

func (t TestFS) Add(path string, data []byte) error {
	return t.add(nil, strings.Split(path, "/"), data)
}

func (t TestFS) Remove(filePath string) error {
	return t.remove(strings.Split(filePath, "/"))
}

func (t TestFS) open(path []string) (ports.ReadWriteFile, error) {
	switch len(path) {
	case 0:
		return nil, fs.ErrNotExist
	case 1:
		f, ok := t.content[path[0]]
		if !ok {
			return nil, fs.ErrNotExist
		}

		if file, ok := f.(ports.ReadWriteFile); ok {
			return file, nil
		}

		return nil, fs.ErrInvalid
	default:
		f, ok := t.content[path[0]]
		if !ok {
			return nil, fs.ErrNotExist
		}
		if tfs, ok := f.(TestFS); ok {
			return tfs.open(path[1:])
		}

		return nil, fs.ErrInvalid
	}
}

func (t TestFS) add(cwd, path []string, data []byte) error {
	if len(path) == 0 {
		return fs.ErrInvalid
	}

	cwd = append(cwd, path[0])

	switch len(path) {
	case 1:
		t.content[path[0]] = &testFile{
			name:    strings.Join(cwd, "/"),
			content: bytes.NewBuffer(data),
		}
		return nil
	default:
		tfs := TestFS{
			relativePath: strings.Join(cwd, "/"),
			content:      map[string]any{},
		}

		t.content[path[0]] = tfs
		return tfs.add(cwd, path[1:], data)
	}
}

func (t TestFS) remove(path []string) error {
	if len(path) == 0 {
		return fs.ErrInvalid
	}

	switch len(path) {
	case 1:
		delete(t.content, path[0])
		return nil
	default:
		f, ok := t.content[path[0]]
		if !ok {
			return fs.ErrNotExist
		}
		if tfs, ok := f.(TestFS); ok {
			return tfs.remove(path[1:])
		}

		return fs.ErrInvalid
	}
}

var _ ports.ReadWriteFile = (*testFile)(nil)

type testFile struct {
	name    string
	content *bytes.Buffer
	reader  *bytes.Reader
}

func (t *testFile) Stat() (fs.FileInfo, error) {
	panic("not supported")
}

func (t *testFile) Read(i []byte) (int, error) {
	if t.reader == nil {
		t.reader = bytes.NewReader(t.content.Bytes())
	}

	return t.reader.Read(i)
}

func (t *testFile) Close() error {
	t.content.Reset()
	t.reader = nil
	return nil
}

func (t *testFile) Write(p []byte) (n int, err error) {
	n, err = t.content.Write(p)
	t.reader = nil
	return
}

func (t *testFile) Seek(offset int64, whence int) (int64, error) {
	if t.reader == nil {
		t.reader = bytes.NewReader(t.content.Bytes())
	}

	return t.reader.Seek(offset, whence)
}

func (t *testFile) WriteString(s string) (int, error) {
	return t.content.WriteString(s)
}

func (t *testFile) Name() string {
	return t.name
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		//nolint:gosec // no need to use cryptographically secure random number generator
		// we are only generating some random file names for testing
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
