# go-git

## Introduction

## Install

See [INSTALL.md](INSTALL.md).

## Getting started

### Init repository to share secret files

```bash
git age init

# or if you want to add some comment to the generated key
git age init -c "My comment"
```

### Add another user to an already prepared repository

__Remarks:__ The repository has to be in a clean state i.e. no changes files.

```bash
git age add-recipient <public key>

# or if you want to add some comment to the added key

git age add-recipient -c "My comment" <public key>
```

`git age add-recipient` will:

1. add the public key to the repository (`.agerecipients` file) 
2. re-encrypt all files with the new set of recipients
3. commit the changes