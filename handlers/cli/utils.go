package cli

import (
	"fmt"
	"io"
	"os"
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
		if err == nil {
			_, err = f.Seek(0, io.SeekStart)
		}
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
