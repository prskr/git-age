package cli

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"io"
	"os"
	"path/filepath"
)

func requireStdin() error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("cannot read from STDIN: %w", err)
	} else if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("cannot read from STDIN: %w", err)
	}

	return nil
}

func copyToTemp(reader io.Reader) (f *os.File, err error) {
	f, err = os.CreateTemp(os.TempDir(), "git-age")
	if err != nil {
		return nil, err
	}

	defer func() {
		_, err = f.Seek(0, io.SeekStart)
	}()

	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}
	}()

	if _, err := io.Copy(f, reader); err != nil {
		return nil, err
	}

	return f, nil
}

func repoRoot(currentDir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
			break
		}

		currentDir = filepath.Dir(currentDir)
		if currentDir == "/" {
			return "", fmt.Errorf("could not find git repository")
		}
	}

	return currentDir, nil
}

func getObjectAtHead(repo *git.Repository, path string) (*object.File, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	return tree.File(path)
}
