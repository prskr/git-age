package fsx

import (
	"errors"
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

func SealingMiddleware(sealer ports.FileSealer, pattern string) SyncerMiddleware {
	return func(filePath string, w io.Writer) (io.WriteCloser, error) {
		match, err := filepath.Match(pattern, filePath)
		if err != nil {
			return nil, err
		}

		if match {
			return sealer.SealFile(w)
		}

		return NopeWriteCloser(w), nil
	}
}

type SyncerMiddleware func(filePath string, w io.Writer) (io.WriteCloser, error)

func NewSyncer(source fs.FS, destination ports.ReadWriteFS, middlewares ...SyncerMiddleware) Syncer {
	return Syncer{
		Source:      source,
		Destination: destination,
		middlewares: middlewares,
	}
}

type Syncer struct {
	Source      fs.FS
	Destination ports.ReadWriteFS
	middlewares []SyncerMiddleware
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
			return err
		}

		defer func() {
			err = errors.Join(err, f.Close())
		}()

		dst, err := s.Destination.OpenRW(path)
		if err != nil {
			return err
		}

		closers := []io.Closer{dst}
		defer func() {
			for i := len(closers) - 1; i >= 0; i-- {
				err = errors.Join(err, closers[i].Close())
			}
		}()

		var w io.WriteCloser = dst

		for _, middleware := range s.middlewares {
			w, err = middleware(path, w)
			if err != nil {
				return err
			}
			closers = append(closers, w)
		}

		_, err = io.Copy(w, f)
		return err
	})
}
