module github.com/prskr/git-age

go 1.25

toolchain go1.25.0

require (
	buf.build/gen/go/git-age/agent/connectrpc/go v1.18.1-20240411154421-ccdd2e6e6f4f.1
	buf.build/gen/go/git-age/agent/protocolbuffers/go v1.36.7-20240411154421-ccdd2e6e6f4f.1
	buf.build/gen/go/grpc/grpc/connectrpc/go v1.18.1-20250429200738-0ee95b84c2c7.1
	buf.build/gen/go/grpc/grpc/protocolbuffers/go v1.36.8-20250429200738-0ee95b84c2c7.1
	connectrpc.com/connect v1.18.1
	connectrpc.com/grpchealth v1.4.0
	filippo.io/age v1.2.1
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/adrg/xdg v0.5.3
	github.com/alecthomas/kong v1.12.1
	github.com/go-git/go-billy/v5 v5.6.2
	github.com/go-git/go-git/v5 v5.16.2
	github.com/lmittmann/tint v1.1.2
	github.com/minio/sha256-simd v1.0.1
	github.com/stretchr/testify v1.10.0
	gopkg.in/ini.v1 v1.67.0
)

replace golang.org/x/crypto => golang.org/x/crypto v0.41.0

tool (
	github.com/go-courier/husky/cmd/husky
	golang.org/x/tools/cmd/goimports
	gotest.tools/gotestsum
	mvdan.cc/gofumpt
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/alecthomas/repr v0.5.1 // indirect
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-courier/husky v1.8.1 // indirect
	github.com/go-courier/semver v1.0.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/onsi/gomega v1.38.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/gotestsum v1.12.3 // indirect
	mvdan.cc/gofumpt v0.8.0 // indirect
	mvdan.cc/sh/v3 v3.12.0 // indirect
)
