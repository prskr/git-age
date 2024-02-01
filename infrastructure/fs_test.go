package infrastructure_test

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prskr/git-age/infrastructure"
)

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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rwfs := infrastructure.NewReadWriteDirFS(t.TempDir())
			f, err := rwfs.TempFile(tt.args.dir, tt.args.pattern)
			if err != nil {
				t.Errorf("TempFile() error = %v", err)
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

func populateTestDirectory(tb testing.TB) (root string) {
	tb.Helper()

	root = tb.TempDir()

	for i := 0; i < 5; i++ {
		dirName := fmt.Sprintf("dir%d", i)
		if err := os.MkdirAll(filepath.Join(root, dirName), 0755); err != nil {
			tb.Fatal(err)
		}

		if i%2 == 0 {
			f, err := os.Create(filepath.Join(root, dirName, fmt.Sprintf("file%d", i)))
			if err != nil {
				tb.Fatal(err)
			}

			if _, err := f.WriteString("hello world"); err != nil {
				tb.Fatal(err)
			}

			_ = f.Close()
		}
	}

	return root
}
