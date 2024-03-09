module github.com/prskr/git-age

go 1.22

toolchain go1.22.1

require (
	filippo.io/age v1.1.1
	github.com/adrg/xdg v0.4.0
	github.com/alecthomas/kong v0.9.0
	github.com/go-git/go-billy/v5 v5.5.0
	github.com/go-git/go-git/v5 v5.11.0
	github.com/lmittmann/tint v1.0.4
	github.com/minio/sha256-simd v1.0.1
	golang.org/x/mod v0.16.0
	gopkg.in/ini.v1 v1.67.0
)

replace (
	github.com/go-git/go-git/v5 => github.com/prskr/go-git/v5 v5.0.0-20240205092825-798d9942c362
	golang.org/x/crypto => golang.org/x/crypto v0.21.0
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.3 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/skeema/knownhosts v1.2.1 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)
