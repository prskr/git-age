package infrastructure_test

import (
	"io"
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
	"github.com/prskr/git-age/internal/testx"
)

func TestNewGitRepositoryFromPath(t *testing.T) {
	t.Parallel()
	type args struct {
		from string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "root of repository",
			args: args{
				from: ".",
			},
		},
		{
			name: "sub-dir of repository",
			args: args{
				from: "dir0",
			},
		},
		{
			name: "no existing repo",
			args: args{
				from: "..",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := populateTestDirectory(t)
			if _, err := git.PlainInit(root, false); err != nil {
				t.Fatalf("failed to initialize git repository: %v", err)
			}

			_, _, err := infrastructure.NewGitRepositoryFromPath(ports.CWD(filepath.Join(root, tt.args.from)))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGitRepositoryFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGitRepository_OpenObjectAtHead(t *testing.T) {
	t.Parallel()
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    func(tb testing.TB, obj *object.File)
		wantErr bool
	}{
		{
			name: "open existing file",
			args: args{
				filePath: "dir0/file0",
			},
			want: func(tb testing.TB, obj *object.File) {
				tb.Helper()

				content, err := obj.Contents()
				if err != nil {
					tb.Fatalf("failed to get file contents: %v", err)
				}

				if content != defaultFileContent {
					tb.Fatalf("unexpected file contents: %s", content)
				}
			},
		},
		{
			name: "open non-existing file",
			args: args{
				filePath: "dir0/file1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, repo := prepareTestRepo(t)
			repoFS := infrastructure.NewReadWriteDirFS(root)

			g := testx.ResultOfA[*infrastructure.GitRepository](t, infrastructure.NewGitRepository, repoFS, repo)

			obj, err := g.OpenObjectAtHead(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenObjectAtHead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				tt.want(t, obj)
			}
		})
	}
}

func TestGitRepository_IsStagingDirty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(tb testing.TB, fs ports.ReadWriteFS, repo *git.Repository)
		want    bool
		wantErr bool
	}{
		{
			name: "staging is clean",
			want: false,
		},
		{
			name: "modified file in working directory",
			want: false,
			setup: func(tb testing.TB, rwfs ports.ReadWriteFS, repo *git.Repository) {
				tb.Helper()

				f := testx.ResultOfA[ports.ReadWriteFile](tb, rwfs.Create, "dir0/file0")

				defer func() {
					if err := f.Close(); err != nil {
						tb.Fatalf("failed to close file: %v", err)
					}
				}()

				testx.ResultOfA[int64](tb, f.Seek, int64(0), io.SeekEnd)
				testx.ResultOfA[int](tb, f.WriteString, "modified")
			},
		},
		{
			name: "staging contains modified file",
			want: true,
			setup: func(tb testing.TB, rwfs ports.ReadWriteFS, repo *git.Repository) {
				tb.Helper()

				f := testx.ResultOfA[ports.ReadWriteFile](tb, rwfs.Create, "dir0/file0")

				defer func() {
					if err := f.Close(); err != nil {
						tb.Fatalf("failed to close file: %v", err)
					}
				}()

				testx.ResultOfA[int64](tb, f.Seek, int64(0), io.SeekEnd)
				testx.ResultOfA[int](tb, f.WriteString, "modified")

				wt := testx.ResultOf(tb, repo.Worktree)
				testx.ResultOfA[plumbing.Hash](tb, wt.Add, filepath.Join("dir0", "file0"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root, repo := prepareTestRepo(t)
			repoFS := infrastructure.NewReadWriteDirFS(root)

			g := testx.ResultOfA[*infrastructure.GitRepository](t, infrastructure.NewGitRepository, repoFS, repo)

			if tt.setup != nil {
				tt.setup(t, repoFS, repo)
			}

			got, err := g.IsStagingDirty()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsStagingDirty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsStagingDirty() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitRepository_WalkAgeFiles(t *testing.T) {
	t.Parallel()

	root, repo := prepareTestRepo(t)
	repoFS := infrastructure.NewReadWriteDirFS(root)

	g := testx.ResultOfA[*infrastructure.GitRepository](t, infrastructure.NewGitRepository, repoFS, repo)

	var discoveredFiles []string

	err := g.WalkAgeFiles(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		discoveredFiles = append(discoveredFiles, path)

		return nil
	})
	if err != nil {
		t.Errorf("WalkAgeFiles() error = %v", err)
		return
	}

	if len(discoveredFiles) != 1 {
		t.Errorf("WalkAgeFiles() discovered files = %v, want 1", discoveredFiles)
	}
}

func prepareTestRepo(tb testing.TB) (root string, repo *git.Repository) {
	tb.Helper()

	root = populateTestDirectory(tb)
	memoryStorer := memory.NewStorage()

	repo = testx.ResultOfA[*git.Repository](tb, git.Init, memoryStorer, osfs.New(root, osfs.WithBoundOS()))

	wt := testx.ResultOf(tb, repo.Worktree)

	if err := wt.AddGlob("dir0/*"); err != nil {
		tb.Fatalf("failed to add files: %v", err)
	}

	commitOptions := &git.CommitOptions{
		Author: &object.Signature{
			Name:  tb.Name(),
			Email: "ci@git-age.io",
			When:  time.Now().UTC(),
		},
	}

	testx.ResultOfA[plumbing.Hash](tb, wt.Commit, "initial commit", commitOptions)

	return root, repo
}
