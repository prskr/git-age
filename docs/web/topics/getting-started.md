# Getting started

## Install hooks

To install the necessary in your Git config run the following command:

```Bash
git age install
```

This will add the _git-age_ clean and smudge filters to your Git config.

## Init a repository to share secret files

```Bash
git age init

# or if you want to add some comment to the generated key
git age init -c "My comment"
```

## Add another user to an already initialized repository

**Remarks:** The repository has to be in a clean state i.e. no changes files.

Alice wants to share the secrets stored in her Git repository with Bob:

1. Bob installs _git-age_ on his machine and configures his global git config
    ```Bash
    git age install
    ```
2. Bob generates a new key pair
    ```Bash
    git age gen-key
    
    # or if you want to add some comment to the generated key
    git age gen-key -c "My comment"
    ```
   the generated private key will be stored automatically in your `keys.txt`
3. Bob sends his public key to Alice
4. Alice adds Bob's public key to her repository
    ```Bash
    git age add-recipient <public key>
    
    # or if you want to add some comment to the added key
    
    git age add-recipient -c "My comment" <public key>
    ```

`git age add-recipient` will:

1. add the public key to the repository (`.agerecipients` file)
2. re-encrypt all files with the new set of recipients
3. commit the changes

As soon as Alice pushed the changes to the remote repository, Bob can pull the changes and decrypt the files.