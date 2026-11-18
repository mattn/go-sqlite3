//go:build !libsqlite3 || sqlite_serialize
// +build !libsqlite3 sqlite_serialize

package sqlite3

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

var deserializeBufOwners sync.Map // uintptr -> *SQLiteConn

func deserializeBufKey(b []byte) uintptr {
	if len(b) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(unsafe.SliceData(b)))
}

func deserializeBufOwner(b []byte) *SQLiteConn {
	if len(b) == 0 {
		return nil
	}
	v, ok := deserializeBufOwners.Load(deserializeBufKey(b))
	if !ok {
		return nil
	}
	return v.(*SQLiteConn)
}

func releaseDeserializeBuf(b []byte, c *SQLiteConn) {
	key := deserializeBufKey(b)
	if key == 0 {
		return
	}
	deserializeBufOwners.CompareAndDelete(key, c)
}

func (c *SQLiteConn) clearDeserializeBufLocked() {
	if c.deserializePinner != nil {
		c.deserializePinner.Unpin()
		c.deserializePinner = nil
	}
	if len(c.deserializeBuf) > 0 {
		releaseDeserializeBuf(c.deserializeBuf, c)
	}
	c.deserializeBuf = nil
}

func (c *SQLiteConn) adoptDeserializeBufLocked(b []byte) error {
	c.clearDeserializeBufLocked()
	if len(b) == 0 {
		c.deserializeBuf = b
		return nil
	}
	if owner := deserializeBufOwner(b); owner != nil && owner != c {
		return fmt.Errorf("deserialize failed: byte slice already active on another connection")
	}
	if c.deserializePinner == nil {
		c.deserializePinner = &runtime.Pinner{}
	} else {
		c.deserializePinner.Unpin()
	}
	c.deserializePinner.Pin(unsafe.SliceData(b))
	deserializeBufOwners.Store(deserializeBufKey(b), c)
	c.deserializeBuf = b
	return nil
}
