// Copyright (C) 2014 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

/*
#include <sqlite3-binding.h>
*/
import "C"

import "unsafe"

//export callbackTrampoline
func callbackTrampoline(ctx *C.sqlite3_context, argc int, argv **C.sqlite3_value) {
	args := (*[1 << 30]*C.sqlite3_value)(unsafe.Pointer(argv))[:argc:argc]
	fi := (*functionInfo)(unsafe.Pointer(C.sqlite3_user_data(ctx)))
	fi.Call(ctx, args)
}
