package sqlite3

import (
	"errors"
)


/*
#cgo CFLAGS: -DSQLITE_OMIT_TRUNCATE_OPTIMIZATION
*/
import "C"
