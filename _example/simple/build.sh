#!/bin/bash

CGO_ENABLED=1 GOARCH=amd64 GOOS=linux CGO_CFLAGS="-DSQLITE_OMIT_LOAD_EXTENSION" go build  -ldflags "-linkmode external -extldflags -static" -a -v --tags "linux sqlite_omit_load_extension" && ldd simple
