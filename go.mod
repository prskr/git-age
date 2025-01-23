module github.com/prskr/git-age

go 1.23

toolchain go1.23.4

require (
	buf.build/gen/go/git-age/agent/connectrpc/go v1.17.0-20240411154421-ccdd2e6e6f4f.1
	buf.build/gen/go/git-age/agent/protocolbuffers/go v1.36.1-20240411154421-ccdd2e6e6f4f.1
	buf.build/gen/go/grpc/grpc/connectrpc/go v1.17.0-20241224201141-c591cbb6df8a.1
	buf.build/gen/go/grpc/grpc/protocolbuffers/go v1.36.1-20241224201141-c591cbb6df8a.1
	connectrpc.com/connect v1.17.0
	connectrpc.com/grpchealth v1.3.0
	filippo.io/age v1.2.1
	github.com/Masterminds/semver/v3 v3.3.1
	github.com/adrg/xdg v0.5.3
	github.com/alecthomas/kong v1.6.0
	github.com/go-git/go-billy/v5 v5.6.2
	github.com/go-git/go-git/v5 v5.13.2
	github.com/lmittmann/tint v1.0.6
	github.com/minio/sha256-simd v1.0.1
	github.com/stretchr/testify v1.10.0
	gopkg.in/ini.v1 v1.67.0
)

replace golang.org/x/crypto => golang.org/x/crypto v0.31.0

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.5 // indirect
	github.com/cloudflare/circl v1.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.3.6 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/skeema/knownhosts v1.3.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
