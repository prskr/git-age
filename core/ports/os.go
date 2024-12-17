package ports

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var ErrUnexpectedEnvVar = errors.New("unexpected environment variable")

type CWD string

func (c CWD) Value() string {
	return string(c)
}

type STDIN io.ReadCloser

type STDOUT io.Writer

func HostEnv() (OSEnv, error) {
	env := make(OSEnv)
	for _, v := range os.Environ() {
		key, value, found := strings.Cut(v, "=")
		if !found {
			return nil, fmt.Errorf("%w: %q", ErrUnexpectedEnvVar, v)
		}
		env[key] = value
	}

	return env, nil
}

func NewOSEnv() OSEnv {
	env := make(OSEnv)
	return env
}

type OSEnv map[string]string

func (e OSEnv) Get(key string) string {
	return e[key]
}
