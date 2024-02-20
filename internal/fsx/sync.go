package fsx

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/prskr/git-age/core/ports"
)

func NopeWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

func SealingMiddleware(sealer ports.FileSealer, pattern string) MiddlewareProvider {
	return func(filePath string, w io.Writer) (io.Writer, error) {
		match, err := filepath.Match(pattern, filePath)
		if err != nil {
			return nil, err
		}

		if match {
			return sealer.SealFile(w)
		}

		return w, nil
	}
}

type MiddlewareProvider func(filePath string, w io.Writer) (io.Writer, error)

var _ io.WriteCloser = (*syncerMiddleware)(nil)

type syncerMiddleware struct {
	writer io.Writer
}

func (s syncerMiddleware) Write(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

func (s syncerMiddleware) Close() error {
	if closer, ok := s.writer.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

func NewSyncer(source fs.FS, destination ports.ReadWriteFS, middlewares ...MiddlewareProvider) Syncer {
	return Syncer{
		Source:      source,
		Destination: destination,
		middlewares: middlewares,
	}
}

type Syncer struct {
	Source      fs.FS
	Destination ports.ReadWriteFS
	middlewares []MiddlewareProvider
}

func (s Syncer) Sync() error {
	return fs.WalkDir(s.Source, ".", func(path string, d fs.DirEntry, walkErr error) (err error) {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return s.Destination.Mkdir(path, true, 0o755)
		}

		f, err := s.Source.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w, path: %s", err, path)
		}

		defer func() {
			err = errors.Join(err, f.Close())
		}()

		dst, err := s.Destination.Create(path, ports.WithTruncate)
		if err != nil {
			return fmt.Errorf("failed to create file: %w, path: %s", err, path)
		}

		var w io.WriteCloser = dst

		defer func() {
			err = errors.Join(err, w.Close())
		}()

		for _, provider := range s.middlewares {
			m, err := provider(path, w)
			if err != nil {
				return err
			}

			w = syncerMiddleware{writer: m}
		}

		_, err = io.Copy(w, f)
		return err
	})
}
