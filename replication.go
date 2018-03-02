package sqlite3

/*
#cgo CFLAGS: -DSQLITE_ENABLE_REPLICATION
#include <string.h>
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>

// SQLite replication hooks
int replicationBegin(void*);
int replicationAbort(void*);
int replicationFrames(void*, int, int, sqlite3_replication_page*, unsigned, int, unsigned);
int replicationUndo(void*);
int replicationEnd(void*);

// SQLite replication implementation
static sqlite3_replication_methods replicationMethods = {
  replicationBegin,
  replicationAbort,
  replicationFrames,
  replicationUndo,
  replicationEnd,
};

// Wrapper around sqlite3_config() for invoking the SQLITE_CONFIG_REPLICATION
// opcode, since there's no way to use C varargs from Go.
static int replicationConfig() {
  return sqlite3_config(SQLITE_CONFIG_REPLICATION, &replicationMethods);
}

// Allocate the given number of replication pages.
static sqlite3_replication_page* replicationPagesAlloc(int nList) {
  return (sqlite3_replication_page*)sqlite3_malloc(sizeof(sqlite3_replication_page) * (nList));
};

// Helper for copying a replication page from Go memory to C memory (pData should
// be an unsafe.Pointer to the first element of a Go slice).
static void replicationPagesFill(sqlite3_replication_page* pList,
  int i, int szPage, void* pData, unsigned flags, unsigned pgno) {
  sqlite3_replication_page* pReplPg = pList + i;
  pReplPg->pBuf = sqlite3_malloc(szPage);
  pReplPg->flags = flags;
  pReplPg->pgno = pgno;
  memcpy(pReplPg->pBuf, pData, szPage);
};

// Release the given number of replication pages.
static void replicationPagesFree(sqlite3_replication_page* pList, int nList) {
  sqlite3_replication_page* pReplPg = pList;
  int i;
  for (i = 0; i < nList; i++) {
     sqlite3_free(pReplPg->pBuf);
     pReplPg += 1;
  }
  sqlite3_free(pList);
};
*/
import "C"
import (
	"reflect"
	"unsafe"
)

func init() {
	// Register the replication implementation
	rc := C.replicationConfig()
	if rc != C.SQLITE_OK {
		panic("failed to configure SQLite replication")
	}
}

// ReplicationMode defines all valid values for replication mode.
type ReplicationMode int

// Available replication modes.
const (
	ReplicationModeNone     = ReplicationMode(int(C.SQLITE_REPLICATION_NONE))
	ReplicationModeLeader   = ReplicationMode(int(C.SQLITE_REPLICATION_LEADER))
	ReplicationModeFollower = ReplicationMode(int(C.SQLITE_REPLICATION_FOLLOWER))
)

// ReplicationPage is just a Go land equivalent of the low-level
// sqlite3_replication_page C type.
//
// FIXME: we should find a cgo-safe way to not copy the memory.
type ReplicationPage struct {
	pBuf  []byte
	flags C.uint
	pgno  C.uint
}

// Fill sets the page attribute using the given parameters.
func (p *ReplicationPage) Fill(data []byte, flags uint16, number uint32) {
	p.pBuf = data
	p.flags = C.uint(flags)
	p.pgno = C.uint(number)
}

// Data returns a pointer to the page data.
func (p *ReplicationPage) Data() []byte {
	return p.pBuf
}

// Flags returns the page flags..
func (p *ReplicationPage) Flags() uint16 {
	return uint16(p.flags)
}

// Number returns the page number.
func (p *ReplicationPage) Number() uint32 {
	return uint32(p.pgno)
}

// NewReplicationPages returns a new slice of n ReplicationPage
// objects, allocated in C memory.
func NewReplicationPages(n int, pageSize int) []ReplicationPage {
	pages := make([]ReplicationPage, n)
	for i := range pages {
		pages[i].pBuf = make([]byte, pageSize)
	}
	return pages
}

// ReplicationFramesParams holds information about a single batch
// of WAL frames that are being dispatched for replication. They map
// to the parameters of the sqlite3_replication_methods.xFrames and
// sqlite3_replication_frames C APIs.
type ReplicationFramesParams struct {
	PageSize  int
	Pages     []ReplicationPage
	Truncate  uint32
	IsCommit  int
	SyncFlags uint8
}

// ReplicationMethods offers a Go-friendly interface around the low level
// sqlite3_replication_methods C type. They are supposed to implement
// application-specific logic in response to replication callbacks
// triggered by sqlite.
type ReplicationMethods interface {

	// Begin a new write transaction. The implementation should check
	// that the connection is eligible for starting a replicated write
	// transaction (e.g. this node is the leader), and perform internal
	// state changes as appropriate.
	Begin(*SQLiteConn) ErrNo

	// Abort a write transaction. The implementation should clear any
	// state previously set by the Begin hook.
	Abort(*SQLiteConn) ErrNo

	// Write new frames to the write-ahead log. The implementation should
	// broadcast this write to other nodes and wait for a quorum.
	Frames(*SQLiteConn, *ReplicationFramesParams) ErrNo

	// Undo a write transaction. The implementation should broadcast
	// this event to other nodes and wait for a quorum. The return code
	// is currently ignored by SQLite.
	Undo(*SQLiteConn) ErrNo

	// End a write transaction. The implementation should update its
	// internal state and be ready for a new transaction.
	End(*SQLiteConn) ErrNo
}

// ReplicationLeader switches this sqlite connection to leader replication
// mode. The given ReplicationMethods instance are hooks for driving the
// execution of the replication in "follower" connections.
func (c *SQLiteConn) ReplicationLeader(methods ReplicationMethods) error {
	handle := newHandle(c, methods)

	rv := C.sqlite3_replication_leader(c.db, replicationSchema, unsafe.Pointer(handle))
	if rv != C.SQLITE_OK {
		return newError(rv)
	}

	return nil
}

// ReplicationFollower switches the given sqlite connection to
// follower replication mode. In this mode no regular operation is
// possible, and the connection should be driven with the
// ReplicationBegin, ReplicationWalFrames, ReplicationCommit and
// ReplicationRollback APIs.
func (c *SQLiteConn) ReplicationFollower() error {
	rv := C.sqlite3_replication_follower(c.db, replicationSchema)
	if rv != C.SQLITE_OK {
		return newError(rv)
	}

	return nil
}

// ReplicationNone switches off replication on the given sqlite connection.
func (c *SQLiteConn) ReplicationNone() error {
	rv := C.sqlite3_replication_none(c.db, replicationSchema)
	if rv != C.SQLITE_OK {
		return newError(rv)
	}
	return nil
}

// ReplicationMode returns the current replication mode of the connection.
func (c *SQLiteConn) ReplicationMode() (ReplicationMode, error) {
	var mode C.int
	rv := C.sqlite3_replication_mode(c.db, replicationSchema, &mode)
	if rv != C.SQLITE_OK {
		return -1, newError(rv)
	}
	return ReplicationMode(mode), nil
}

// ReplicationFrames writes the given batch of frames to the write-ahead log
// linked to the given connection. This should be called with a "follower"
// connection, meant to replicate the "leader" one.
func ReplicationFrames(conn *SQLiteConn, begin bool, params *ReplicationFramesParams) error {
	// Convert to C types
	db := conn.db
	isBegin := C.int(0)
	if begin {
		isBegin = C.int(1)
	}
	szPage := C.int(params.PageSize)
	nList := C.int(len(params.Pages))
	nTruncate := C.uint(params.Truncate)
	isCommit := C.int(params.IsCommit)
	syncFlags := C.int(params.SyncFlags)

	// FIXME: avoid the copy
	pList := C.replicationPagesAlloc(nList)
	defer C.replicationPagesFree(pList, nList)

	for i := range params.Pages {
		C.replicationPagesFill(
			pList, C.int(i), szPage, unsafe.Pointer(reflect.ValueOf(params.Pages[i].pBuf).Pointer()),
			params.Pages[i].flags, params.Pages[i].pgno)
	}

	rc := C.sqlite3_replication_frames(
		db, replicationSchema, isBegin, szPage, nList, pList, nTruncate, isCommit, syncFlags)
	if rc != C.SQLITE_OK {
		return newError(rc)
	}
	return nil
}

// ReplicationUndo rollbacks a write transaction in the given sqlite
// connection. This should be called with a "follower" connection,
// meant to replicate the "leader" one.
func ReplicationUndo(conn *SQLiteConn) error {
	rc := C.sqlite3_replication_undo(conn.db, replicationSchema)
	if rc != C.SQLITE_OK {
		return newError(rc)
	}
	return nil
}

// NoopReplicationMethods returns a new instance of a ReplicationMethods
// implementation whose hooks do nothing.
func NoopReplicationMethods() ReplicationMethods {
	return &noopReplicationMethods{}
}

type noopReplicationMethods struct{}

func (m *noopReplicationMethods) Begin(conn *SQLiteConn) ErrNo {
	return 0
}

func (m *noopReplicationMethods) Abort(conn *SQLiteConn) ErrNo {
	return 0
}

func (m *noopReplicationMethods) Frames(conn *SQLiteConn, params *ReplicationFramesParams) ErrNo {
	return 0
}

func (m *noopReplicationMethods) Undo(conn *SQLiteConn) ErrNo {
	return 0
}

func (m *noopReplicationMethods) End(conn *SQLiteConn) ErrNo {
	return 0
}

//export replicationBegin
//
// Hook implementing sqlite3_replication_methods->xBegin
func replicationBegin(pArg unsafe.Pointer) C.int {
	handle := lookupHandleVal(uintptr(pArg))
	conn := handle.db
	methods := handle.val.(ReplicationMethods)
	return C.int(methods.Begin(conn))
}

//export replicationAbort
//
// Hook implementing sqlite3_replication_methods->xAbort
func replicationAbort(pArg unsafe.Pointer) C.int {
	handle := lookupHandleVal(uintptr(pArg))
	conn := handle.db
	methods := handle.val.(ReplicationMethods)
	return C.int(methods.Abort(conn))
}

//export replicationFrames
//
// Hook implementing sqlite3_replication_methods->xFrames
func replicationFrames(pArg unsafe.Pointer, szPage C.int, nList C.int, pList *C.sqlite3_replication_page, nTruncate C.uint, isCommit C.int, syncFlags C.uint) C.int {
	pages := NewReplicationPages(int(nList), int(szPage))
	size := unsafe.Sizeof(C.sqlite3_replication_page{})
	for i := range pages {
		pPage := (*C.sqlite3_replication_page)(
			unsafe.Pointer((uintptr(unsafe.Pointer(pList)) + size*uintptr(i))))
		C.memcpy(unsafe.Pointer(reflect.ValueOf(pages[i].pBuf).Pointer()), pPage.pBuf, C.size_t(szPage))
		pages[i].pgno = pPage.pgno
		pages[i].flags = pPage.flags
	}

	params := &ReplicationFramesParams{
		PageSize:  int(szPage),
		Pages:     pages,
		Truncate:  uint32(nTruncate),
		IsCommit:  int(isCommit),
		SyncFlags: uint8(syncFlags),
	}

	handle := lookupHandleVal(uintptr(pArg))
	conn := handle.db
	methods := handle.val.(ReplicationMethods)

	return C.int(methods.Frames(conn, params))
}

//export replicationUndo
//
// Hook implementing sqlite3_replication_methods->xUndo
func replicationUndo(pArg unsafe.Pointer) C.int {
	handle := lookupHandleVal(uintptr(pArg))
	conn := handle.db
	methods := handle.val.(ReplicationMethods)
	return C.int(methods.Undo(conn))
}

//export replicationEnd
//
// Hook implementing sqlite3_replication_methods->xEnd
func replicationEnd(pArg unsafe.Pointer) C.int {
	handle := lookupHandleVal(uintptr(pArg))
	conn := handle.db
	methods := handle.val.(ReplicationMethods)
	return C.int(methods.End(conn))
}

// Hard-coded main schema name.
//
// TODO: support replicating also attached databases.
var replicationSchema = C.CString("main")
