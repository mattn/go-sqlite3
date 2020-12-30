// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
*/
import "C"

// ErrNoMask is mask code.
const ErrNoMask C.int = 0xff

// Error return error message that is extended code.
func (err ErrNoExtended) Error() string {
	return Error{Code: ErrNo(C.int(err) & ErrNoMask), ExtendedCode: err}.Error()
}

func (err Error) Error() string {
	var str string
	if err.err != "" {
		str = err.err
	} else {
		str = C.GoString(C.sqlite3_errstr(C.int(err.Code)))
	}
	if err.SystemErrno != 0 {
		str += ": " + err.SystemErrno.Error()
	}
	return str
}
