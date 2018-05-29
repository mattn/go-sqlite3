go-sqlite3
==========

[![GoDoc Reference](https://godoc.org/github.com/mattn/go-sqlite3?status.svg)](http://godoc.org/github.com/mattn/go-sqlite3)
[![Build Status](https://travis-ci.org/mattn/go-sqlite3.svg?branch=master)](https://travis-ci.org/mattn/go-sqlite3)
[![Coverage Status](https://coveralls.io/repos/mattn/go-sqlite3/badge.svg?branch=master)](https://coveralls.io/r/mattn/go-sqlite3?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattn/go-sqlite3)](https://goreportcard.com/report/github.com/mattn/go-sqlite3)

# Description

sqlite3 driver conforming to the built-in database/sql interface

Supported Golang version:
- 1.8.x
- 1.9.x
- 1.10.x

### Overview

- [Installation](#installation)
- [API Reference](#api-reference)
- [Connection String](#connection-string)
- [Features](#features)
- [Compilation](#compilation)
  - [Android](#android)
  - [ARM](#arm)
  - [Cross Compile](#cross-compile)
  - [Google Cloud Platform](#google-cloud-platform)
  - [Linux](#linux)
    - [Alpine](#alpine)
    - [Fedora](#fedora)
    - [Ubuntu](#ubuntu)
  - [Mac OSX](#mac-osx)
  - [Windows](#windows)
  - [Errors](#errors)
- [FAQ](#faq)
- [License](#license)

# Installation

This package can be installed with the go get command:

    go get github.com/mattn/go-sqlite3

_go-sqlite3_ is *cgo* package.
If you want to build your app using go-sqlite3, you need gcc.
However, after you have built and installed _go-sqlite3_ with `go install github.com/mattn/go-sqlite3` (which requires gcc), you can build your app without relying on gcc in future.

***Important: because this is a `CGO` enabled package you are required to set the environment variable `CGO_ENABLED=1` and have a `gcc` compile present within your path.***

# API Reference

API documentation can be found here: http://godoc.org/github.com/mattn/go-sqlite3

Examples can be found under the [examples](./_example) directory

# Connection String

When creating a new SQLite database or connection to an existing one, with the file name additional options can be given.
This is also known as a DSN string. (Data Source Name).

Options are append after the filename of the SQLite database.
The database filename and options are seperated by an `?` (Question Mark).

This also applies when using an in-memory database instead of a file.

Options can be given using the following format: `KEYWORD=VALUE` and multiple options can be combined with the `&` ampersand.

This library supports dsn options of SQLite itself and provides additional options.

Boolean values can be one of:
* `0` `no` `false` `off`
* `1` `yes` `true` `on`

| Name | Key | Value(s) | Description |
|------|-----|----------|-------------|
| Auto Vacuum | `_vacuum` | <ul><li>`0` \| `none`</li><li>`1` \| `full`</li><li>`2` \| `incremental`</li></ul> | For more information see [PRAGMA auto_vacuum](https://www.sqlite.org/pragma.html#pragma_auto_vacuum) |
| Busy Timeout | `_busy_timeout` \| `_timeout` | `int` | Specify value for sqlite3_busy_timeout. For more information see [PRAGMA busy_timeout](https://www.sqlite.org/pragma.html#pragma_busy_timeout) |
| Case Sensitive LIKE | `_cslike` | `boolean` | For more information see [PRAGMA case_sensitive_like](https://www.sqlite.org/pragma.html#pragma_case_sensitive_like) |
| Defer Foreign Keys | `_defer_foreign_keys` \| `_defer_fk` | `boolean` | For more information see [PRAGMA defer_foreign_keys](https://www.sqlite.org/pragma.html#pragma_defer_foreign_keys) |
| Foreign Keys | `_foreign_keys` \| `_fk` | `boolean` | For more information see [PRAGMA foreign_keys](https://www.sqlite.org/pragma.html#pragma_foreign_keys) |
| Ignore CHECK Constraints | `_ignore_check_constraints` | `boolean` | For more information see [PRAGMA ignore_check_constraints](https://www.sqlite.org/pragma.html#pragma_ignore_check_constraints) |
| Mode | `mode` | <ul><li>ro</li><li>rw</li><li>rwc</li><li>memory</li></ul> | Access Mode of the database. For more information see [SQLite Open](https://www.sqlite.org/c3ref/open.html) |
| Mutex Locking | `_mutex` | <ul><li>no</li><li>full</li></ul> | Specify mutex mode. |
| Recursive Triggers | `_recursive_triggers` \| `_rt` | `boolean` | For more information see [PRAGMA recursive_triggers](https://www.sqlite.org/pragma.html#pragma_recursive_triggers) |
| Shared-Cache Mode | `cache` | <ul><li>shared</li><li>private</li></ul> | Set cache mode for more information see [sqlite.org](https://www.sqlite.org/sharedcache.html) |
| Time Zone Location | `_loc` | auto | Specify location of time format. |
| Transaction Lock | `_txlock` | <ul><li>immediate</li><li>deferred</li><li>exclusive</li></ul> | Specify locking behavior for transactions. |

## DSN Examples

```
file:test.db?cache=shared&mode=memory
```

# Features

This package allows additional configuration of features available within SQLite3 to be enabled or disabled by golang build constraints also known as build `tags`.

[Click here for more information about build tags / constraints.](https://golang.org/pkg/go/build/#hdr-Build_Constraints)

### Usage

If you wish to build this library with additional extensions / features.
Use the following command.

```bash
go build --tags "<FEATURE>"
```

For available features see the extension list.
When using multiple build tags, all the different tags should be space delimted.

Example:

```bash
go build --tags "icu json1 fts5 secure_delete"
```

### Feature / Extension List

| Extension | Build Tag | Description |
|-----------|-----------|-------------|
| Additional Statistics | sqlite_stat4 | This option adds additional logic to the ANALYZE command and to the query planner that can help SQLite to chose a better query plan under certain situations. The ANALYZE command is enhanced to collect histogram data from all columns of every index and store that data in the sqlite_stat4 table.<br><br>The query planner will then use the histogram data to help it make better index choices. The downside of this compile-time option is that it violates the query planner stability guarantee making it more difficult to ensure consistent performance in mass-produced applications.<br><br>SQLITE_ENABLE_STAT4 is an enhancement of SQLITE_ENABLE_STAT3. STAT3 only recorded histogram data for the left-most column of each index whereas the STAT4 enhancement records histogram data from all columns of each index.<br><br>The SQLITE_ENABLE_STAT3 compile-time option is a no-op and is ignored if the SQLITE_ENABLE_STAT4 compile-time option is used |
| Allow URI Authority | sqlite_allow_uri_authority | URI filenames normally throws an error if the authority section is not either empty or "localhost".<br><br>However, if SQLite is compiled with the SQLITE_ALLOW_URI_AUTHORITY compile-time option, then the URI is converted into a Uniform Naming Convention (UNC) filename and passed down to the underlying operating system that way |
| App Armor | sqlite_app_armor | When defined, this C-preprocessor macro activates extra code that attempts to detect misuse of the SQLite API, such as passing in NULL pointers to required parameters or using objects after they have been destroyed. <br><br>App Armor is not available under `Windows`. |
| Disable Load Extensions | sqlite_omit_load_extension | Loading of external extensions is enabled by default.<br><br>To disable extension loading add the build tag `sqlite_omit_load_extension`. |
| Foreign Keys | sqlite_foreign_keys | This macro determines whether enforcement of foreign key constraints is enabled or disabled by default for new database connections.<br><br>Each database connection can always turn enforcement of foreign key constraints on and off and run-time using the foreign_keys pragma.<br><br>Enforcement of foreign key constraints is normally off by default, but if this compile-time parameter is set to 1, enforcement of foreign key constraints will be on by default | 
| Full Auto Vacuum | sqlite_vacuum_full | Set the default auto vacuum to full |
| Incremental Auto Vacuum | sqlite_vacuum_incr | Set the default auto vacuum to incremental |
| Full Text Search Engine | sqlite_fts5 | When this option is defined in the amalgamation, versions 5 of the full-text search engine (fts5) is added to the build automatically |
|  International Components for Unicode | sqlite_icu | This option causes the International Components for Unicode or "ICU" extension to SQLite to be added to the build |
| Introspect PRAGMAS | sqlite_introspect | This option adds some extra PRAGMA statements. <ul><li>PRAGMA function_list</li><li>PRAGMA module_list</li><li>PRAGMA pragma_list</li></ul> |
| JSON SQL Functions | sqlite_json | When this option is defined in the amalgamation, the JSON SQL functions are added to the build automatically |
| Secure Delete | sqlite_secure_delete | This compile-time option changes the default setting of the secure_delete pragma.<br><br>When this option is not used, secure_delete defaults to off. When this option is present, secure_delete defaults to on.<br><br>The secure_delete setting causes deleted content to be overwritten with zeros. There is a small performance penalty since additional I/O must occur.<br><br>On the other hand, secure_delete can prevent fragments of sensitive information from lingering in unused parts of the database file after it has been deleted. See the documentation on the secure_delete pragma for additional information |
| Tracing / Debug | sqlite_trace | Activate trace functions |

# Compilation

This package requires `CGO_ENABLED=1` ennvironment variable if not set by default, and the presence of the `gcc` compiler.

If you need to add additional CFLAGS or LDFLAGS to the build command, and do not want to modify this package. Then this can be achieved by  using the `CGO_CFLAGS` and `CGO_LDFLAGS` environment variables.

## Android

This package can be compiled for android.
Compile with:

```bash
go build --tags "android"
```

For more information see [#201](https://github.com/mattn/go-sqlite3/issues/201)

# ARM

To compile for `ARM` use the following environment.

```bash
env CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ \
    CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 \
    go build -v 
```

Additional information:
- [#242](https://github.com/mattn/go-sqlite3/issues/242)
- [#504](https://github.com/mattn/go-sqlite3/issues/504)

# Cross Compile

This library can be cross-compiled.

In some cases you are required to the `CC` environment variable with the cross compiler.

Additional information:
- [#491](https://github.com/mattn/go-sqlite3/issues/491)
- [#560](https://github.com/mattn/go-sqlite3/issues/560)

# Google Cloud Platform

Building on GCP is not possible because `Google Cloud Platform does not allow `gcc` to be executed.

Please work only with compiled final binaries.

## Linux

To compile this package on Linux you must install the development tools for your linux distribution.

To compile under linux use the build tag `linux`.

```bash
go build --tags "linux"
```

If you wish to link directly to libsqlite3 then you can use the `libsqlite3` build tag.

```
go build --tags "libsqlite3 linux"
```

### Alpine

When building in an `alpine` container run the following command before building.

```
apk add --update gcc musl-dev
```

### Fedora

```bash
sudo yum groupinstall "Development Tools" "Development Libraries"
```

### Ubuntu

```bash
sudo apt-get install build-essential
```

## Mac OSX

OSX should have all the tools present to compile this package, if not install XCode this will add all the developers tools.

Required dependency

```bash
brew install sqlite3
```

For OSX there is an additional package install which is required if you whish to build the `icu` extension.

This additional package can be installed with `homebrew`.

```bash
brew upgrade icu4c
```

To compile for Mac OSX.

```bash
go build --tags "darwin"
```

If you wish to link directly to libsqlite3 then you can use the `libsqlite3` build tag.

```
go build --tags "libsqlite3 darwin"
```

Additional information:
- [#206](https://github.com/mattn/go-sqlite3/issues/206)
- [#404](https://github.com/mattn/go-sqlite3/issues/404)

## Windows

To compile this package on Windows OS you must have the `gcc` compiler installed.

1) Install a Windows `gcc` toolchain.
2) Add the `bin` folders to the Windows path if the installer did not do this by default.
3) Open a terminal for the TDM-GCC toolchain, can be found in the Windows Start menu.
4) Navigate to your project folder and run the `go build ...` command for this package.

For example the TDM-GCC Toolchain can be found [here](ttps://sourceforge.net/projects/tdm-gcc/).

## Errors

- Compile error: `can not be used when making a shared object; recompile with -fPIC`

    When receiving a compile time error referencing recompile with `-FPIC` then you
    are probably using a hardend system.

    You can copile the library on a hardend system with the following command.

    ```bash
    go build -ldflags '-extldflags=-fno-PIC'
    ```

    More details see [#120](https://github.com/mattn/go-sqlite3/issues/120)

- Can't build go-sqlite3 on windows 64bit.

    > Probably, you are using go 1.0, go1.0 has a problem when it comes to compiling/linking on windows 64bit.
    > See: [#27](https://github.com/mattn/go-sqlite3/issues/27)

- `go get github.com/mattn/go-sqlite3` throws compilation error.

    `gcc` throws: `internal compiler error`

    Remove the download repository from your disk and try re-install with:

    ```bash
    go install github.com/mattn/go-sqlite3
    ```

# FAQ

- Getting insert error while query is opened.

    > You can pass some arguments into the connection string, for example, a URI.
    > See: [#39](https://github.com/mattn/go-sqlite3/issues/39)

- Do you want to cross compile? mingw on Linux or Mac?

    > See: [#106](https://github.com/mattn/go-sqlite3/issues/106)
    > See also: http://www.limitlessfx.com/cross-compile-golang-app-for-windows-from-linux.html

- Want to get time.Time with current locale

    Use `_loc=auto` in SQLite3 filename schema like `file:foo.db?_loc=auto`.

- Can I use this in multiple routines concurrently?

    Yes for readonly. But, No for writable. See [#50](https://github.com/mattn/go-sqlite3/issues/50), [#51](https://github.com/mattn/go-sqlite3/issues/51), [#209](https://github.com/mattn/go-sqlite3/issues/209), [#274](https://github.com/mattn/go-sqlite3/issues/274).

- Why is it racy if I use a `sql.Open("sqlite3", ":memory:")` database?

    Each connection to :memory: opens a brand new in-memory sql database, so if
    the stdlib's sql engine happens to open another connection and you've only
    specified ":memory:", that connection will see a brand new database. A
    workaround is to use "file::memory:?mode=memory&cache=shared". Every
    connection to this string will point to the same in-memory database. See
    [#204](https://github.com/mattn/go-sqlite3/issues/204) for more info.

- Reading from database with large amount of goroutines fails on OSX.

    OS X limits OS-wide to not have more than 1000 files open simultaneously by default.

    For more information see [#289](https://github.com/mattn/go-sqlite3/issues/289)

- Trying to execure a `.` (dot) command throws an error.

    Error: `Error: near ".": syntax error`
    Dot command are part of SQLite3 CLI not of this library.

    You need to implement the feature or call the sqlite3 cli.

    More infomation see [#305](https://github.com/mattn/go-sqlite3/issues/305)

# License

MIT: http://mattn.mit-license.org/2012

sqlite3-binding.c, sqlite3-binding.h, sqlite3ext.h

The -binding suffix was added to avoid build failures under gccgo.

In this repository, those files are an amalgamation of code that was copied from SQLite3. The license of that code is the same as the license of SQLite3.

# Author

Yasuhiro Matsumoto (a.k.a mattn)
