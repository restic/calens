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

# Creating a Changelog

To create a changlog.md file using calens, change into a repository containing a `changelog`
folder similar to that of [restic](https://github.com/restic/restic/tree/master/changelog).
Then run `calens -o changlog.md` from within the repository.

When done, open the created changelog to see the generated changelog.

Run `calens --help` for more options.
