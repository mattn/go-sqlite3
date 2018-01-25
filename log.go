package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif

// The log error callback to pass to sqlite3, implemented in Go.
extern void logCallback(void*, int, char*);

// Conveniences to call sqlite3_config, since CGO doesn't play well
// with varargs.
static int logConfigure() {
  return sqlite3_config(SQLITE_CONFIG_LOG, logCallback, 0);
}
static int logUnconfigure() {
  return sqlite3_config(SQLITE_CONFIG_LOG, 0, 0);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func init() {
	logSafeConfigure()
}

// LogFunc is a function that can process SQLite logging events.
type LogFunc func(int, string)

// LogConfig configures SQLite to report errors using the given function. If
// nil is passed, then logging is turned off. It returns whatever previous
// value was set (either a function or nil).
func LogConfig(f LogFunc) LogFunc {
	p := logFunc
	logFunc = f
	return p
}

//export logCallback
func logCallback(pArg unsafe.Pointer, iErrCode C.int, zMsg *C.char) {
	if logFunc == nil {
		return
	}
	logFunc(int(iErrCode), C.GoString(zMsg))
}

// Call the C logConfigure binding and panic in case of error.
func logSafeConfigure() {
	if rc := C.logConfigure(); rc != 0 {
		msg := errorString(Error{Code: ErrNo(rc)})
		panic(fmt.Sprintf("failed to initialize SQLite logging: %s (%d)", msg, rc))
	}
}

var logFunc LogFunc
