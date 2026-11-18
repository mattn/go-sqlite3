//go:build !libsqlite3 || sqlite_serialize
// +build !libsqlite3 sqlite_serialize

package sqlite3

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// deserializeGlobalMu serializes buffer ownership checks and adoption across
// connections for the same backing array.
var deserializeGlobalMu sync.Mutex

// deserializeBufOwners maps a buffer key to the owning connection ID (uintptr).
// Values are opaque IDs, not *SQLiteConn, so the map does not retain connections.
var deserializeBufOwners sync.Map // uintptr -> uintptr

func connDeserializeID(c *SQLiteConn) uintptr {
	return uintptr(unsafe.Pointer(c))
}

func deserializeBufKey(b []byte) uintptr {
	if len(b) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(unsafe.SliceData(b)))
}

func deserializeBufOwnerID(b []byte) (uintptr, bool) {
	if len(b) == 0 {
		return 0, false
	}
	v, ok := deserializeBufOwners.Load(deserializeBufKey(b))
	if !ok {
		return 0, false
	}
	return v.(uintptr), true
}

func releaseDeserializeBuf(b []byte, c *SQLiteConn) {
	key := deserializeBufKey(b)
	if key == 0 {
		return
	}
	id := connDeserializeID(c)
	deserializeBufOwners.CompareAndDelete(key, id)
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

func (c *SQLiteConn) beginDeserializeBufLocked(b []byte) error {
	c.clearDeserializeBufLocked()
	if len(b) == 0 {
		c.deserializeBuf = b
		return nil
	}
	if ownerID, ok := deserializeBufOwnerID(b); ok && ownerID != connDeserializeID(c) {
		return fmt.Errorf("deserialize failed: byte slice already active on another connection")
	}
	if c.deserializePinner == nil {
		c.deserializePinner = &runtime.Pinner{}
	} else {
		c.deserializePinner.Unpin()
	}
	c.deserializePinner.Pin(unsafe.SliceData(b))
	c.deserializeBuf = b
	return nil
}

func (c *SQLiteConn) commitDeserializeBufLocked(b []byte) {
	if len(b) > 0 {
		deserializeBufOwners.Store(deserializeBufKey(b), connDeserializeID(c))
	}
}

func (c *SQLiteConn) rollbackDeserializeBufLocked() {
	if c.deserializePinner != nil {
		c.deserializePinner.Unpin()
		c.deserializePinner = nil
	}
	c.deserializeBuf = nil
}
