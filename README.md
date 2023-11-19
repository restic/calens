# calens

`calens` is a changelog generator that nicely prints a changelog file from
a directory with versioned subdirs containing files. A sample directory can be
seen in [the restic repository](https://github.com/restic/restic/tree/master/changelog)

# Installation

To install and run calens locally, run the following commands:
```
git clone https://github.com/restic/calens.git 
cd calens
CGO_ENABLED=0 go install
```

Note that you need to to update the local repo and rerun installation when a new calens release becomes avialable.
```
cd calens
git pull --rebase origin master
CGO_ENABLED=0 go install
```

# Testing

To test calens and create a `test-changlog.md` file, change into a repository containing
a `changelog` folder similar to that of [restic](https://github.com/restic/restic/tree/master/changelog).
Then run `calens -o test-changlog.md` from within the repository.

When done, open the created changelog with a browser to see the rendered content.

Run `calens --help` for more options.
