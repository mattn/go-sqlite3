# Go-SQLite3

[![GoDoc Reference](https://godoc.org/github.com/mattn/go-sqlite3?status.svg)](http://godoc.org/github.com/mattn/go-sqlite3)
[![Build Status](https://travis-ci.org/mattn/go-sqlite3.svg?branch=master)](https://travis-ci.org/mattn/go-sqlite3)
[![Coverage Status](https://coveralls.io/repos/mattn/go-sqlite3/badge.svg?branch=master)](https://coveralls.io/r/mattn/go-sqlite3?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattn/go-sqlite3)](https://goreportcard.com/report/github.com/mattn/go-sqlite3)

**Current Version: 2.0.0**

**Please note that version 2.0.0 is not backwards compatible**

## Documentation

[More documentation is available in our Wiki](https://github.com/mattn/go-sqlite3/wiki)

## Description

sqlite3 driver conforming to the built-in database/sql interface

Supported Golang version:
- 1.9.x
- 1.10.x

[This package follows the official Golang Release Policy.](https://golang.org/doc/devel/release.html#policy)

## Requirements

* Go: 1.9 or higher.
* GCC (Go-SQLite3 is a `CGO` package)

_go-sqlite3_ is *cgo* package.
If you want to build your app using go-sqlite3, you need gcc.
However, after you have built and installed _go-sqlite3_ with `go install github.com/mattn/go-sqlite3` (which requires gcc), you can build your app without relying on gcc in future.

**Important: because this is a `CGO` enabled package you are required to set the environment variable `CGO_ENABLED=1` and have a `gcc` compile present within your path.**

## Installation

```bash
go get github.com/mattn/go-sqlite3/driver
```

## Usage

`Go-SQLite3 is an implementation of Go's database/sql/driver interface. You only need to import the driver and can use the full database/sql API then.

Use `sqlite3` as `drivername` and a valid [DSN](https://github.com/mattn/go-sqlite3/wiki/DSN) as `dataSourceName`:

```go
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3/driver"
)

db, err := sql.Open("sqlite3", "file:test.db")
```

[Examples are available in our Wiki](https://github.com/mattn/go-sqlite3/wiki/Examples)

# License

MIT: http://mattn.mit-license.org/2018

sqlite3-binding.c, sqlite3-binding.h, sqlite3ext.h

The -binding suffix was added to avoid build failures under gccgo.

In this repository, those files are an amalgamation of code that was copied from SQLite3. 
The license of that code is the same as the license of SQLite3.
