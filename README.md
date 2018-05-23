go-sqlite3
==========

[![GoDoc Reference](https://godoc.org/github.com/mattn/go-sqlite3?status.svg)](http://godoc.org/github.com/mattn/go-sqlite3)
[![Build Status](https://travis-ci.org/mattn/go-sqlite3.svg?branch=master)](https://travis-ci.org/mattn/go-sqlite3)
[![Coverage Status](https://coveralls.io/repos/mattn/go-sqlite3/badge.svg?branch=master)](https://coveralls.io/r/mattn/go-sqlite3?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattn/go-sqlite3)](https://goreportcard.com/report/github.com/mattn/go-sqlite3)

# Description

sqlite3 driver conforming to the built-in database/sql interface

### Overview

- [Installation](#installation)
- [API Reference](#api-reference)
- [Features](#features)
- [FAQ](#faq)
- [License](#license)

# Installation

This package can be installed with the go get command:

    go get github.com/mattn/go-sqlite3

_go-sqlite3_ is *cgo* package.
If you want to build your app using go-sqlite3, you need gcc.
However, after you have built and installed _go-sqlite3_ with `go install github.com/mattn/go-sqlite3` (which requires gcc), you can build your app without relying on gcc in future.

# API Reference

API documentation can be found here: http://godoc.org/github.com/mattn/go-sqlite3

Examples can be found under the [examples](./_example) directory

# Features

This package allows additional configuration of features available within SQLite3 to be enabled or disabled by golang build constraints also known as build `tags`.

[Click here for more information about build tags / constraints.](https://golang.org/pkg/go/build/#hdr-Build_Constraints)

| Extension | Build Tag | Description |
|-----------|-----------|-------------|
| Additional Statistics | stat4 | This option adds additional logic to the ANALYZE command and to the query planner that can help SQLite to chose a better query plan under certain situations. The ANALYZE command is enhanced to collect histogram data from all columns of every index and store that data in the sqlite_stat4 table.<br><br>The query planner will then use the histogram data to help it make better index choices. The downside of this compile-time option is that it violates the query planner stability guarantee making it more difficult to ensure consistent performance in mass-produced applications.<br><br>SQLITE_ENABLE_STAT4 is an enhancement of SQLITE_ENABLE_STAT3. STAT3 only recorded histogram data for the left-most column of each index whereas the STAT4 enhancement records histogram data from all columns of each index.<br><br>The SQLITE_ENABLE_STAT3 compile-time option is a no-op and is ignored if the SQLITE_ENABLE_STAT4 compile-time option is used |
| Allow URI Authority | allow_uri_authority | URI filenames normally throws an error if the authority section is not either empty or "localhost".<br><br>However, if SQLite is compiled with the SQLITE_ALLOW_URI_AUTHORITY compile-time option, then the URI is converted into a Uniform Naming Convention (UNC) filename and passed down to the underlying operating system that way |
| App Armor | app_armor | When defined, this C-preprocessor macro activates extra code that attempts to detect misuse of the SQLite API, such as passing in NULL pointers to required parameters or using objects after they have been destroyed. <br><br>App Armor is not available under `Windows`. |
| Disable Load Extensions | sqlite_omit_load_extension | Loading of external extensions is enabled by default.<br><br>To disable extension loading add the build tag `sqlite_omit_load_extension`. |
| Foreign Keys | foreign_keys | This macro determines whether enforcement of foreign key constraints is enabled or disabled by default for new database connections.<br><br>Each database connection can always turn enforcement of foreign key constraints on and off and run-time using the foreign_keys pragma.<br><br>Enforcement of foreign key constraints is normally off by default, but if this compile-time parameter is set to 1, enforcement of foreign key constraints will be on by default | 
| Full Auto Vacuum | vacuum_full | Set the default auto vacuum to full |
| Incremental Auto Vacuum | vacuum_incr | Set the default auto vacuum to incremental |
| Full Text Search Engine | fts5 | When this option is defined in the amalgamation, versions 5 of the full-text search engine (fts5) is added to the build automatically |
|  International Components for Unicode | icu | This option causes the International Components for Unicode or "ICU" extension to SQLite to be added to the build |
| Introspect PRAGMAS | introspect | This option adds some extra PRAGMA statements. <ul><li>PRAGMA function_list</li><li>PRAGMA module_list</li><li>PRAGMA pragma_list</li></ul> |
| JSON SQL Functions | json1 | When this option is defined in the amalgamation, the JSON SQL functions are added to the build automatically |
| Secure Delete | secure_delete | This compile-time option changes the default setting of the secure_delete pragma.<br><br>When this option is not used, secure_delete defaults to off. When this option is present, secure_delete defaults to on.<br><br>The secure_delete setting causes deleted content to be overwritten with zeros. There is a small performance penalty since additional I/O must occur.<br><br>On the other hand, secure_delete can prevent fragments of sensitive information from lingering in unused parts of the database file after it has been deleted. See the documentation on the secure_delete pragma for additional information |

# FAQ

- Want to build go-sqlite3 with libsqlite3 on my linux.

    Use `go build --tags "libsqlite3 linux"`

- Want to build go-sqlite3 with libsqlite3 on OS X.

    Install sqlite3 from homebrew: `brew install sqlite3`

    Use `go build --tags "libsqlite3 darwin"`

- Want to build go-sqlite3 with additional extensions / features.

    Use `go build --tags "<FEATURE>"`

    When using multiple build tags, all the different tags should be space delimted.

    For available features / extensions see [Features](#features).

    Example building multiple features: `go build --tags "icu json1 fts5 secure_delete"`

- Can't build go-sqlite3 on windows 64bit.

    > Probably, you are using go 1.0, go1.0 has a problem when it comes to compiling/linking on windows 64bit.
    > See: [#27](https://github.com/mattn/go-sqlite3/issues/27)

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

# License

MIT: http://mattn.mit-license.org/2012

sqlite3-binding.c, sqlite3-binding.h, sqlite3ext.h

The -binding suffix was added to avoid build failures under gccgo.

In this repository, those files are an amalgamation of code that was copied from SQLite3. The license of that code is the same as the license of SQLite3.

# Author

Yasuhiro Matsumoto (a.k.a mattn)
