package infrastructure

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitattributes"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/prskr/git-age/core/ports"
	"io/fs"
	"path/filepath"
	"strings"
)

var (
	_ ports.RepoStater       = (*GitRepository)(nil)
	_ ports.RepoWalker       = (*GitRepository)(nil)
	_ ports.Comitter         = (*GitRepository)(nil)
	_ ports.HeadObjectOpener = (*GitRepository)(nil)
)

func NewGitRepository(repoFS fs.FS, repository *git.Repository) (*GitRepository, error) {
	wt, err := repository.Worktree()
	if err != nil {
		return nil, err
	}

	return &GitRepository{
		RepoFS:     repoFS,
		Repository: repository,
		Worktree:   wt,
	}, err
}

type GitRepository struct {
	RepoFS     fs.FS
	Repository *git.Repository
	Worktree   *git.Worktree
}

func (g GitRepository) OpenObjectAtHead(filePath string) (*object.File, error) {
	head, err := g.Repository.Head()
	if err != nil {
		return nil, err
	}

	commit, err := g.Repository.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	return tree.File(filePath)
}

func (g GitRepository) StageFile(path string) error {
	_, err := g.Worktree.Add(path)
	return err
}

func (g GitRepository) Commit(message string) error {
	_, err := g.Worktree.Commit(message, new(git.CommitOptions))
	return err
}

func (g GitRepository) WalkAgeFiles(onMatch fs.WalkDirFunc) error {
	matchAttrs, err := gitattributes.ReadPatterns(g.Worktree.Filesystem, nil)
	if err != nil {
		return err
	}

	matcher := gitattributes.NewMatcher(matchAttrs)
	wantedAttributes := []string{"age"}

	return fs.WalkDir(g.RepoFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path == ".git" {
				return fs.SkipDir
			}
			return nil
		}

		_, matched := matcher.Match(strings.Split(path, string([]rune{filepath.Separator})), wantedAttributes)
		if matched {
			if err := onMatch(path, d, err); err != nil {
				return err
			}
		}

		return nil
	})
}

func (g GitRepository) IsDirty() (bool, error) {
	status, err := g.Worktree.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean(), nil
}
