#!/bin/bash

CGO_ENABLED=1 GOARCH=amd64 GOOS=linux CGO_CFLAGS="-DSQLITE_HAS_CODEC -DSQLITE_TEMP_STORE=2" GOLDFLAGS="-linkmode external -extldflags -static" CGO_LDFLAGS="/usr/lib/x86_64-linux-gnu/libcrypto.a" go build -v --tags "linux"
