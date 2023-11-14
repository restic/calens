# calens

`calens` is a changelog generator that nicely prints a changelog file from
a directory with versioned subdirs containing files. A sample directory can be
seen in [the restic repository](https://github.com/restic/restic/tree/master/changelog)

# Installation

To install and run calens locally, clone this repo and run from the command line `go install`.
The installation process downloads all required libraries. With some OS, you may need to install
as prerequisite, if the OS provides it like Ubuntu, the `build essentials` package.
The installation process responds in case which part is missing. 

# Testing

To test calens and create a `test-changlog.md` file you can open with a browser, change into
a cloned repo satisfying the above requirements and run `calens -o test-changlog.md`.
When done, open the created changelog with a browser to see the rendered content.

Run `calens --help` for more options.
