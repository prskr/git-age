package ports

import (
	"errors"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"io/fs"
	"os"
)

var ErrNoGlobalConfig = errors.New("no global git config found")

type RepoStater interface {
	IsDirty() (bool, error)
}

type RepoWalker interface {
	WalkAgeFiles(walkFunc fs.WalkDirFunc) error
}

type Comitter interface {
	AddFile(path string) error
	Commit(message string) error
}

type HeadObjectOpener interface {
	OpenObjectAtHead(objectPath string) (*object.File, error)
}

type GitRepository interface {
	RepoStater
	RepoWalker
	Comitter
	HeadObjectOpener
}

func GlobalGitConfigPath() (string, error) {
	configPaths, err := config.Paths(config.GlobalScope)
	if err != nil {
		return "", err
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return "", err
		} else {
			return path, nil
		}
	}

	return "", ErrNoGlobalConfig
}
