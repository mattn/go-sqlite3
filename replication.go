package sqlite3

/*
#include <string.h>
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>

// SQLite replication hooks
extern int replicationBegin(void *pArg);
extern int replicationWalFrames(void *pArg, int szPage, int nList, sqlite3_replication_page *pList, unsigned nTruncate, int isCommit, unsigned sync_flags);
extern int replicationUndo(void *pArg);
extern int replicationEnd(void *pArg);
extern int replicationCheckpoint(void *pArg, int eMode, int *pnLog, int *pnCkpt);

// SQLite replication implementation
static sqlite3_replication_methods replicationMethods = {
  replicationBegin,
  replicationWalFrames,
  replicationUndo,
  replicationEnd,
  replicationCheckpoint
};

static int replicationLeader(sqlite3 *db) {
  return sqlite3_replication_leader(db, "main", &replicationMethods, db);
};

static sqlite3_replication_page* replicationPagesAlloc(int nList) {
  return (sqlite3_replication_page*)sqlite3_malloc(sizeof(sqlite3_replication_page) * (nList));
};

static void replicationPagesFill(sqlite3_replication_page* pList,
  int i, int szPage, void* pData, unsigned flags, unsigned pgno) {
  sqlite3_replication_page* pReplPg = pList + i;
  pReplPg->pBuf = sqlite3_malloc(szPage);
  pReplPg->flags = flags;
  pReplPg->pgno = pgno;
  memcpy(pReplPg->pBuf, pData, szPage);
};

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
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// Replication defines all valid values for replication mode.
type Replication int

// Available replication modes.
const (
	ReplicationModeNone     = Replication(int(C.SQLITE_REPLICATION_NONE))
	ReplicationModeLeader   = Replication(int(C.SQLITE_REPLICATION_LEADER))
	ReplicationModeFollower = Replication(int(C.SQLITE_REPLICATION_FOLLOWER))
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

// ReplicationWalFramesParams holds information about a single batch
// of WAL frames that are being dispatched for replication. They map
// to the parameters of the sqlite3_replication_methods.xWalFrames and
// sqlite3_replication_wal_frames C APIs.
type ReplicationWalFramesParams struct {
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

	// Begin a new write transaction. The implementation should
	// eventually trigger the execution of the Begin function
	// contained in this package, invoking it once for every follower
	// connections that the application wants to use to replicate
	// the leader.
	Begin(*SQLiteConn) ErrNo

	// Write new frames to the write-ahead log. The implementation should
	// eventually trigger the execution of the WalFrames function
	// contained in this package, invoking it once for every follower
	// connections that the application wants to use to replicate
	// the leader.
	WalFrames(*SQLiteConn, *ReplicationWalFramesParams) ErrNo

	// Undo a write transaction. The implementation should
	// eventually trigger the execution of the Rollback function
	// contained in this package, invoking it once for every follower
	// connections that the application wants to use to replicate
	// the leader.
	Undo(*SQLiteConn) ErrNo

	// Commit a write transaction. The implementation should
	// eventually trigger the execution of the Commit function
	// contained in this package, invoking it once for every follower
	// connections that the application wants to use to replicate
	// the leader.
	End(*SQLiteConn) ErrNo

	// Checkpoint the current WAL frames.
	Checkpoint(*SQLiteConn, WalCheckpointMode, *int, *int) ErrNo
}

// ReplicationLeader switches the given sqlite connection to leader
// replication mode. The given ReplicationMethods instance are hooks for
// driving the execution of the replication in "follower" connections.
func ReplicationLeader(conn *SQLiteConn, methods ReplicationMethods) error {
	db := conn.db
	instance := findMethodsInstance(db)

	if instance != nil {
		return fmt.Errorf("leader replication already enabled for this connection")
	}

	if rc := C.replicationLeader(db); rc != C.SQLITE_OK {
		return newError(rc)
	}

	registerMethodsInstance(conn, methods)

	return nil
}

// ReplicationFollower switches the given sqlite connection to
// follower replication mode. In this mode no regular operation is
// possible, and the connection should be driven with the
// ReplicationBegin, ReplicationWalFrames, ReplicationCommit and
// ReplicationRollback APIs.
func ReplicationFollower(conn *SQLiteConn) error {
	db := conn.db
	zSchema := C.CString("main")
	defer C.free(unsafe.Pointer(zSchema))

	if rc := C.sqlite3_replication_follower(db, zSchema); rc != C.SQLITE_OK {
		return newError(rc)
	}
	return nil
}

// ReplicationNone switches off replication on the given sqlite
// connection. Note that only leader replication mode can be switched
// off. For follower replication you need to close the connection.
func ReplicationNone(conn *SQLiteConn) (ReplicationMethods, error) {
	db := conn.db
	instance := findMethodsInstance(db)

	// Switch leader replication off
	if instance == nil {
		return nil, fmt.Errorf("leader replication is not enabled for this connection")
	}

	zSchema := C.CString("main")
	defer C.free(unsafe.Pointer(zSchema))

	if rc := C.sqlite3_replication_none(db, zSchema); rc != C.SQLITE_OK {
		return nil, newError(rc)
	}

	unregisterMethodsInstance(conn)
	return instance.methods, nil
}

// ReplicationMode returns the current replication mode of the given
// transaction.
func ReplicationMode(conn *SQLiteConn) (Replication, error) {
	db := conn.db
	var mode C.int

	zSchema := C.CString("main")
	defer C.free(unsafe.Pointer(zSchema))

	if rc := C.sqlite3_replication_mode(db, zSchema, &mode); rc != C.SQLITE_OK {
		return 0, newError(rc)
	}
	return Replication(int(mode)), nil
}

// ReplicationBegin starts a new write transaction in the given sqlite
// connection. This should be called against a "follower" connection,
// meant to replicate the "leader" one.
func ReplicationBegin(conn *SQLiteConn) error {
	zSchema := C.CString("main")
	defer C.free(unsafe.Pointer(zSchema))

	if rc := C.sqlite3_replication_begin(conn.db, zSchema); rc != C.SQLITE_OK {
		return newError(rc)
	}
	return nil
}

// ReplicationWalFrames writes the given batch of frames to the
// write-ahead log linked to the given connection. This should be
// called with a "follower" connection, meant to replicate the
// "leader" one.
func ReplicationWalFrames(conn *SQLiteConn, params *ReplicationWalFramesParams) error {
	// Convert to C types
	db := conn.db
	zSchema := C.CString("main")
	defer C.free(unsafe.Pointer(zSchema))
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

	if rc := C.sqlite3_replication_wal_frames(
		db, zSchema, szPage, nList, pList, nTruncate, isCommit, syncFlags); rc != C.SQLITE_OK {
		return newError(rc)
	}
	return nil
}

// ReplicationUndo rollbacks a write transaction in the given sqlite
// connection. This should be called with a "follower" connection,
// meant to replicate the "leader" one.
func ReplicationUndo(conn *SQLiteConn) error {
	zSchema := C.CString("main")
	defer C.free(unsafe.Pointer(zSchema))

	if rc := C.sqlite3_replication_undo(conn.db, zSchema); rc != C.SQLITE_OK {
		return newError(rc)
	}
	return nil
}

// ReplicationEnd finishes a write transaction in the given sqlite
// connection. This should be called with a "follower" connection,
// meant to replicate the "leader" one.
func ReplicationEnd(conn *SQLiteConn) error {
	zSchema := C.CString("main")
	defer C.free(unsafe.Pointer(zSchema))

	if rc := C.sqlite3_replication_end(conn.db, zSchema); rc != C.SQLITE_OK {
		return newError(rc)
	}
	return nil
}

// ReplicationCheckpoint checkpoints the current WAL.
func ReplicationCheckpoint(conn *SQLiteConn, mode WalCheckpointMode, log *int, ckpt *int) error {
	// Convert to C types
	db := conn.db
	zSchema := C.CString("main")
	defer C.free(unsafe.Pointer(zSchema))

	eMode := C.int(mode)
	pnLog := (*C.int)(unsafe.Pointer(log))
	pnCkpt := (*C.int)(unsafe.Pointer(ckpt))

	if rc := C.sqlite3_replication_checkpoint(db, zSchema, eMode, pnLog, pnCkpt); rc != C.SQLITE_OK {
		return newError(rc)
	}
	return nil
}

// ReplicationAutoCheckpoint can be used to enable autocheckpoint of
// the replicated WAL.
func ReplicationAutoCheckpoint(conn *SQLiteConn, n int) {

	callback := func(arg unsafe.Pointer, conn *SQLiteConn, database string, frame int) error {
		if frame >= n {
			_, _, err := WalCheckpointV2(conn, WalCheckpointTruncate)
			if err != nil {
				return err
			}
		}
		return nil
	}

	WalHook(conn, callback, nil)
}

// PassthroughReplicationMethods returns a new instance of a ReplicationMethods
// implementation whose hooks just invoke the relevant replication APIs against
// whatever conn.SQLiteConn connection was used to at registration time. If any
// error is hit in the hooks, a panic is raised. This should be only used as
// helper in unit tests.
func PassthroughReplicationMethods() ReplicationMethods {
	return &passthroughReplicationMethods{}
}

type passthroughReplicationMethods struct {
	followers []*SQLiteConn
}

func (m *passthroughReplicationMethods) Begin(conn *SQLiteConn) ErrNo {
	return m.check(ReplicationBegin(conn))
}

func (m *passthroughReplicationMethods) WalFrames(conn *SQLiteConn, params *ReplicationWalFramesParams) ErrNo {
	return m.check(ReplicationWalFrames(conn, params))
}

func (m *passthroughReplicationMethods) Undo(conn *SQLiteConn) ErrNo {
	return m.check(ReplicationUndo(conn))
}

func (m *passthroughReplicationMethods) End(conn *SQLiteConn) ErrNo {
	return m.check(ReplicationEnd(conn))
}

func (m *passthroughReplicationMethods) Checkpoint(conn *SQLiteConn, mode WalCheckpointMode, log *int, ckpt *int) ErrNo {
	return m.check(ReplicationCheckpoint(conn, mode, log, ckpt))
}

// Check that the given error is nil, and return 0 if so. Otherwise, panic out.
func (m *passthroughReplicationMethods) check(err error) ErrNo {
	if err != nil {
		panic(err)
	}
	return 0
}

//export replicationBegin
//
// Hook implementing sqlite3_replication_methods->xBegin
func replicationBegin(pArg unsafe.Pointer) C.int {
	instance := mustFindMethodsInstance((*C.sqlite3)(pArg))
	return C.int(instance.methods.Begin(instance.conn))
}

//export replicationWalFrames
//
// Hook implementing sqlite3_replication_methods->xWalFrames
func replicationWalFrames(pArg unsafe.Pointer, szPage C.int, nList C.int, pList *C.sqlite3_replication_page, nTruncate C.uint, isCommit C.int, syncFlags C.uint) C.int {
	instance := mustFindMethodsInstance((*C.sqlite3)(pArg))

	pages := NewReplicationPages(int(nList), int(szPage))
	size := unsafe.Sizeof(C.sqlite3_replication_page{})
	for i := range pages {
		pPage := (*C.sqlite3_replication_page)(
			unsafe.Pointer((uintptr(unsafe.Pointer(pList)) + size*uintptr(i))))
		C.memcpy(unsafe.Pointer(reflect.ValueOf(pages[i].pBuf).Pointer()), pPage.pBuf, C.size_t(szPage))
		pages[i].pgno = pPage.pgno
		pages[i].flags = pPage.flags
	}

	params := &ReplicationWalFramesParams{
		PageSize:  int(szPage),
		Pages:     pages,
		Truncate:  uint32(nTruncate),
		IsCommit:  int(isCommit),
		SyncFlags: uint8(syncFlags),
	}

	return C.int(instance.methods.WalFrames(instance.conn, params))
}

//export replicationUndo
//
// Hook implementing sqlite3_replication_methods->xUndo
func replicationUndo(pArg unsafe.Pointer) C.int {
	instance := mustFindMethodsInstance((*C.sqlite3)(pArg))
	return C.int(instance.methods.Undo(instance.conn))
}

//export replicationEnd
//
// Hook implementing sqlite3_replication_methods->xEnd
func replicationEnd(pArg unsafe.Pointer) C.int {
	instance := mustFindMethodsInstance((*C.sqlite3)(pArg))
	return C.int(instance.methods.End(instance.conn))
}

//export replicationCheckpoint
//
// Hook implementing sqlite3_replication_methods->xCheckpoint
func replicationCheckpoint(pArg unsafe.Pointer, eMode C.int, pnLog *C.int, pnCkpt *C.int) C.int {
	instance := mustFindMethodsInstance((*C.sqlite3)(pArg))
	return C.int(instance.methods.Checkpoint(
		instance.conn, WalCheckpointMode(eMode), (*int)(unsafe.Pointer(pnLog)),
		(*int)(unsafe.Pointer(pnCkpt)),
	))
}

// Information about a C.sqlite3_replication_methods instance that was
// allocated by this module using C.sqlite3_malloc in ReplicationLeader,
// and its associated ReplicationMethods hooks in Go land.
type replicationMethodsInstance struct {
	conn    *SQLiteConn
	methods ReplicationMethods
}

// A registry for tracking instances of C.sqlite3_replication_methods
// allocated by this module.
//
// Each key is a pointer to a C.sqlite3 db object, and each associated
// value a replicationMethodsInstance holding a pointer to the C callbacks
// were registered against that db, and the Methods object that those
// callbacks will be dispatched to.
var replicationMethodsRegistry = make(map[uintptr]*replicationMethodsInstance)

// Serialize access to the methods registry
var replicationMethodsMutex sync.RWMutex

// Register a new replication for the given connection and methods.
func registerMethodsInstance(conn *SQLiteConn, methods ReplicationMethods) {
	replicationMethodsMutex.Lock()
	defer replicationMethodsMutex.Unlock()

	pointer := uintptr(unsafe.Pointer(conn.db))
	replicationMethodsRegistry[pointer] = &replicationMethodsInstance{
		conn:    conn,
		methods: methods,
	}
}

// Unregister a previously registered replication
func unregisterMethodsInstance(conn *SQLiteConn) {
	replicationMethodsMutex.Lock()
	defer replicationMethodsMutex.Unlock()

	pointer := uintptr(unsafe.Pointer(conn.db))
	delete(replicationMethodsRegistry, pointer)
}

// Find the replication methods instance for the given database.
func findMethodsInstance(db *C.sqlite3) *replicationMethodsInstance {
	replicationMethodsMutex.RLock()
	defer replicationMethodsMutex.RUnlock()

	pointer := uintptr(unsafe.Pointer(db))
	return replicationMethodsRegistry[pointer]
}

// Find the replication methods instance for the given database and
// ensure they are valid.
func mustFindMethodsInstance(db *C.sqlite3) *replicationMethodsInstance {
	instance := findMethodsInstance(db)
	if instance == nil {
		// Something really bad happened if we have lost track of this
		// connection.
		panic("replication hooks not found")
	}
	return instance
}
