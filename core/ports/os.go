package ports

import "io"

type CWD string

func (c CWD) Value() string {
	return string(c)
}

type STDIN io.ReadCloser

type STDOUT io.Writer
