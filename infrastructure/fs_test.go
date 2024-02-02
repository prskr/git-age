package infrastructure_test

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/internal/testx"

	"github.com/prskr/git-age/infrastructure"
)

const defaultFileContent = "Hello, world!"

func TestReadWriteDirFS_Rename(t *testing.T) {
	t.Parallel()

	type args struct {
		oldPath string
		newPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "rename existing file",
			args: args{
				oldPath: path.Join("dir0", "file0"),
				newPath: path.Join("dir0", "file1"),
			},
		},
		{
			name: "rename non-existing file",
			args: args{
				oldPath: path.Join("dir1", "file0"),
				newPath: path.Join("dir0", "file1"),
			},
			wantErr: true,
		},
		{
			name: "rename directory",
			args: args{
				oldPath: "dir0",
				newPath: "dir10",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := populateTestDirectory(t)

			rwfs := infrastructure.NewReadWriteDirFS(root)

			if err := rwfs.Rename(tt.args.oldPath, tt.args.newPath); (err != nil) != tt.wantErr {
				t.Errorf("Rename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// skip further checks
			if tt.wantErr {
				return
			}

			if _, err := fs.Stat(rwfs, tt.args.oldPath); !os.IsNotExist(err) {
				t.Fatalf("expected file to not exist, got %v", err)
			}

			if f, err := rwfs.Open(tt.args.newPath); err != nil {
				t.Fatalf("expected file to exist, got %v", err)
			} else {
				_ = f.Close()
			}
		})
	}
}

func TestReadWriteDirFS_TempFile(t *testing.T) {
	t.Parallel()
	type args struct {
		dir     string
		pattern string
	}
	tests := []struct {
		name string
		args args
		want func(tb testing.TB, name string)
	}{
		{
			name: "create temp file in root",
			args: args{
				dir:     "",
				pattern: "tempfile",
			},
			want: func(tb testing.TB, name string) {
				tb.Helper()

				if !strings.HasPrefix(name, "tempfile") {
					tb.Fatalf("expected file name to start with tempfile, got %s", name)
				}
			},
		},
		{
			name: "create temp file in sub directory",
			args: args{
				dir:     "dir1",
				pattern: "tempfile",
			},
			want: func(tb testing.TB, name string) {
				tb.Helper()

				if !strings.HasPrefix(name, path.Join("dir1", "tempfile")) {
					tb.Fatalf("expected file name to start with tempfile, got %s", name)
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()

			if err := os.MkdirAll(filepath.Join(dir, "dir1"), 0o755); err != nil {
				t.Fatal(err)
			}

			rwfs := infrastructure.NewReadWriteDirFS(dir)
			f, err := rwfs.TempFile(tt.args.dir, tt.args.pattern)
			if err != nil {
				t.Errorf("TempFile() error = %v", err)
				return
			}

			t.Cleanup(func() {
				_ = f.Close()
			})

			if tt.want != nil {
				tt.want(t, f.Name())
			}
		})
	}
}

func TestReadWriteDirFS_Remove(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		pathToDelete string
		wantErr      bool
	}{
		{
			name:         "remove existing file",
			pathToDelete: path.Join("dir0", "file0"),
		},
		{
			name:         "remove non-existing file",
			pathToDelete: path.Join("dir1", "file1"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := populateTestDirectory(t)

			rwfs := infrastructure.NewReadWriteDirFS(root)

			if err := rwfs.Remove(tt.pathToDelete); (err != nil) != tt.wantErr {
				t.Fatal(err)
			}

			if _, err := fs.Stat(rwfs, tt.pathToDelete); !os.IsNotExist(err) {
				t.Fatalf("expected file to not exist, got %v", err)
			}
		})
	}
}

func TestReadWriteDirFS_OpenRW(t *testing.T) {
	t.Parallel()
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    func(tb testing.TB, f ports.ReadWriteFile)
	}{
		{
			name: "create new file",
			args: args{
				filePath: path.Join("dir0", "file1"),
			},
		},
		{
			name: "open existing file",
			args: args{
				filePath: path.Join("dir0", "file0"),
			},
			want: func(tb testing.TB, f ports.ReadWriteFile) {
				tb.Helper()
				content, err := io.ReadAll(f)
				if err != nil {
					tb.Fatal(err)
				}

				if string(content) != defaultFileContent {
					tb.Fatalf("expected file content to be hello world, got %s", string(content))
				}
			},
		},
		{
			name: "append to existing file",
			args: args{
				filePath: path.Join("dir0", "file0"),
			},
			want: func(tb testing.TB, f ports.ReadWriteFile) {
				tb.Helper()
				_, err := f.Write([]byte("hello"))
				if err != nil {
					tb.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := populateTestDirectory(t)
			f := infrastructure.NewReadWriteDirFS(root)

			got, err := f.OpenRW(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenRW() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				tt.want(t, got)
			}

			t.Cleanup(func() {
				_ = got.Close()
			})
		})
	}
}

func TestReadWriteDirFS_Walk(t *testing.T) {
	t.Parallel()
	root := populateTestDirectory(t)
	rwfs := infrastructure.NewReadWriteDirFS(root)

	err := fs.WalkDir(rwfs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		f, err := rwfs.OpenRW(path)
		if err != nil {
			return err
		}

		return f.Close()
	})
	if err != nil {
		t.Errorf("Walk() error = %v", err)
	}
}

func populateTestDirectory(tb testing.TB) (root string) {
	tb.Helper()

	root = tb.TempDir()
	wd := testx.ResultOf(tb, os.Getwd)

	srcFS := os.DirFS(filepath.Join(wd, "testdata"))
	err := fs.WalkDir(srcFS, ".", func(path string, d fs.DirEntry, walkErr error) (err error) {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return os.MkdirAll(filepath.Join(root, path), 0o755)
		}

		f, err := srcFS.Open(path)
		if err != nil {
			return err
		}

		defer func() {
			err = errors.Join(err, f.Close())
		}()

		dst, err := os.Create(filepath.Join(root, path))
		if err != nil {
			return err
		}

		defer func() {
			err = errors.Join(err, dst.Close())
		}()

		_, err = io.Copy(dst, f)
		return err
	})
	if err != nil {
		tb.Fatalf("failed to populate test directory: %v", err)
	}

	return root
}
