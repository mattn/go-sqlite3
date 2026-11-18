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

func (c *SQLiteConn) checkDeserializeBufOwnershipLocked(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if ownerID, ok := deserializeBufOwnerID(b); ok && ownerID != connDeserializeID(c) {
		return fmt.Errorf("deserialize failed: byte slice already active on another connection")
	}
	return nil
}

func pinDeserializeSlice(pinner *runtime.Pinner, b []byte) {
	if len(b) == 0 {
		return
	}
	pinner.Pin(unsafe.SliceData(b))
}

// installDeserializeBufLocked replaces any previously retained buffer after
// sqlite3_deserialize succeeds.
func (c *SQLiteConn) installDeserializeBufLocked(b []byte) {
	c.clearDeserializeBufLocked()
	if len(b) == 0 {
		c.deserializeBuf = b
		return
	}
	if c.deserializePinner == nil {
		c.deserializePinner = &runtime.Pinner{}
	}
	c.deserializePinner.Pin(unsafe.SliceData(b))
	deserializeBufOwners.Store(deserializeBufKey(b), connDeserializeID(c))
	c.deserializeBuf = b
}
