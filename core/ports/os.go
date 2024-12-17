package ports

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type CWD string

func (c CWD) Value() string {
	return string(c)
}

type STDIN io.ReadCloser

type STDOUT io.Writer

func HostEnv() OSEnv {
	env := make(OSEnv)
	for _, v := range os.Environ() {
		key, value, found := strings.Cut(v, "=")
		if !found {
			panic(fmt.Sprintf("unexpected environment variable %s", v))
		}
		env[key] = value
	}

	return env
}

func NewOSEnv() OSEnv {
	env := make(OSEnv)
	return env
}

type OSEnv map[string]string

func (e OSEnv) Get(key string) string {
	return e[key]
}
