module github.com/prskr/git-age

go 1.26.2

require (
	buf.build/gen/go/git-age/agent/connectrpc/go v1.19.1-20240411154421-ccdd2e6e6f4f.2
	buf.build/gen/go/git-age/agent/protocolbuffers/go v1.36.11-20240411154421-ccdd2e6e6f4f.1
	buf.build/gen/go/grpc/grpc/connectrpc/go v1.19.1-20260331211127-1730f7242d0f.2
	buf.build/gen/go/grpc/grpc/protocolbuffers/go v1.36.11-20260331211127-1730f7242d0f.1
	connectrpc.com/connect v1.19.1
	connectrpc.com/grpchealth v1.4.0
	filippo.io/age v1.3.1
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/adrg/xdg v0.5.3
	github.com/alecthomas/kong v1.15.0
	github.com/go-git/go-billy/v5 v5.8.0
	github.com/go-git/go-git/v5 v5.17.2
	github.com/lmittmann/tint v1.1.3
	github.com/minio/sha256-simd v1.0.1
	github.com/stretchr/testify v1.11.1
	gopkg.in/ini.v1 v1.67.1
)

replace golang.org/x/crypto => golang.org/x/crypto v0.50.0

tool (
	golang.org/x/tools/cmd/goimports
	gotest.tools/gotestsum
	mvdan.cc/gofumpt
)

require (
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/hpke v0.4.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.4.1 // indirect
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.6.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.21 // indirect
	github.com/onsi/gomega v1.38.0 // indirect
	github.com/pjbgf/sha1cd v0.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/skeema/knownhosts v1.3.2 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/telemetry v0.0.0-20260414141209-fac6e1c83189 // indirect
	golang.org/x/term v0.42.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/gotestsum v1.13.0 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)
