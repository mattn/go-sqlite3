## How to compile

```bash
cd ${GOPATH}/src/github.com/mattn/go-sqlite3/examples/mod_regexp
make all
```

## How to run

The OS has to be able to find the compiled extension.
In normal cases this library (.so) should be added to the LD_LIBRARY_PATH.

Run run from current compiled directory.

```bash
LD_LIBRARY_PATH=$(pwd) ./extension
```