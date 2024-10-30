# Installations

### GitHub releases

See the [releases page](https://github.com/prskr/git-age/blob/main/prskr/git-age/releases/latest) for the latest release.
You can download the binary for your platform from there.

## MacOS

## Homebrew

There's a tap for _git-age_ available at [prskr/the-prancing-package](https://github.com/prskr/the-prancing-package).

```Bash
brew install prskr/the-prancing-package/git-age
```

or

```Bash
brew tap prskr/the-prancing-package

brew install git-age
```

## Linux

### RPM - Fedora, CentOS, RHEL, SuSE, ...

#### DNF

```Bash
# Import the GPG key
curl https://api.github.com/users/prskr/gpg_keys | jq -r '.[] | select (.key_id=="1A80DDB584AF7DA7") | .raw_key' > /tmp/prskr.gpg
sudo rpm --import /tmp/prskr.gpg

dnf config-manager --add-repo https://code.icb4dc0.de/api/packages/prskr/rpm.repo

dnf install git-age
```

#### Zypper

```
zypper addrepo https://code.icb4dc0.de/api/packages/prskr/rpm.repo

zypper install git-age
```

### DEB - Debian, Ubuntu, Mint, PopOS!, ...

```Bash
sudo curl https://code.icb4dc0.de/api/packages/prskr/debian/repository.key -o /etc/apt/trusted.gpg.d/forgejo-prskr.asc

# distribution is currently only bookworm - but should work for other debian based distributions as well
echo "deb https://code.icb4dc0.de/api/packages/prskr/debian bookworm main" | sudo tee -a /etc/apt/sources.list.d/forgejo.list
sudo curl https://code.icb4dc0.de/api/packages/prskr/debian/repository.key -o /etc/apt/trusted.gpg.d/forgejo-prskr.asc

sudo apt update

sudo apt install git-age
```

### Arch Linux

As part of the release process an AUR (Arch User Repository) package is published.
You can either install it via the default mechanism:

```Bash
git clone https://aur.archlinux.org/git-age-bin.git /tmp/git-age-bin
cd /tmp/git-age-bin
makepkg -si
cd
rm -rf /tmp/git-age-bin
```

or of course with your AUR wrapper like [`yay`](https://github.com/Jguer/yay)

```Bash
yay -S git-age-bin
```

## Windows

To install _git-age_ on Windows, you can use [winget](https://learn.microsoft.com/en-us/windows/package-manager/winget/) or [scoop](https://scoop.sh/).

### Winget

```Bash
winget install --id=prskr.git-age
```

### Scoop

I maintain a bucket for _git-age_ at [prskr/scoop-the-prancing-package](https://github.com/prskr/scoop-the-prancing-package).
To add the bucket and install _git-age_, run the following commands:

```Bash
scoop bucket add the-prancing-package https://github.com/prskr/scoop-the-prancing-package
scoop install git-age
```