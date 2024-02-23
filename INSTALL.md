# Installation of `git-age`

## Go install

If you have a current Go installation, you can install `git-age` with:

```sh
go install github.com/prskr/git-age@latest
```

## GitHub releases

See [the releases page](/prskr/git-age/releases/latest) for the latest release.
You can download the binary for your platform from there.

## Homebrew

There's a tap for `git-age` available at [prskr/the-prancing-package](https://github.com/prskr/the-prancing-package).

```bash
brew install prskr/the-prancing-package/git-age
```

or

```bash
brew tap prskr/the-prancing-package

brew install git-age
```

## Linux

### RPM - Fedora, CentOS, RHEL, SUSE

#### DNF

```bash
# Import the GPG key
curl https://api.github.com/users/prskr/gpg_keys | jq -r '.[] | select (.key_id=="1A80DDB584AF7DA7") | .raw_key' > /tmp/prskr.gpg
sudo rpm --import /tmp/prskr.gpg

dnf config-manager --add-repo https://code.icb4dc0.de/api/packages/prskr/rpm.repo

dnf install git-age
```

#### Zypper

```bash
zypper addrepo https://code.icb4dc0.de/api/packages/prskr/rpm.repo

zypper install git-age
```

### DEB - Debian, Ubuntu

```bash
sudo curl https://code.icb4dc0.de/api/packages/prskr/debian/repository.key -o /etc/apt/trusted.gpg.d/forgejo-prskr.asc

# distribution is currently only bookworm - but should work for other debian based distributions as well
echo "deb https://code.icb4dc0.de/api/packages/prskr/debian bookworm main" | sudo tee -a /etc/apt/sources.list.d/forgejo.list
sudo curl https://code.icb4dc0.de/api/packages/prskr/debian/repository.key -o /etc/apt/trusted.gpg.d/forgejo-prskr.asc

sudo apt update

sudo apt install git-age
```

## Windows

To install `git-age` on Windows, you can use [Scoop](https://scoop.sh/).

### Scoop

I maintain a bucket for `git-age` at [prskr/scoop-the-prancing-package](https://github.com/prskr/scoop-the-prancing-package).
To add the bucket and install `git-age`, run the following commands:

```powershell
scoop bucket add the-prancing-package https://github.com/prskr/scoop-the-prancing-package
scoop install git-age
```