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
		keyValue := strings.Split(v, "=")
		if len(keyValue) != 2 {
			panic(fmt.Sprintf("unexpected environment variable %s", keyValue))
		}
		env[keyValue[0]] = keyValue[1]
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
