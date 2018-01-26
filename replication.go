package sqlite3

/*
#include <string.h>
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif

// SQLite replication hooks
extern int replicationBegin(void *pArg);
extern int replicationWalFrames(void *pArg, int szPage, int nList, sqlite3_replication_page *pList, unsigned nTruncate, int isCommit, unsigned sync_flags);
extern int replicationUndo(void *pArg);
extern int replicationEnd(void *pArg);
extern int replicationCheckpoint(void *pArg, int eMode, int *pnLog, int *pnCkpt);
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

// ReplicationPage is just a Go land alias of the low-level C type.
type ReplicationPage C.sqlite3_replication_page

// Fill sets the page attribute using the given parameters.
func (p *ReplicationPage) Fill(data []byte, flags uint16, number uint32) {
	C.memcpy(p.pBuf, unsafe.Pointer(reflect.ValueOf(data).Pointer()), C.size_t(len(data)))
	p.flags = C.uint(flags)
	p.pgno = C.uint(number)
}

// Data returns a pointer to the page data.
func (p *ReplicationPage) Data() unsafe.Pointer {
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
	size := C.int(unsafe.Sizeof(C.sqlite3_replication_page{}))
	pList := unsafe.Pointer(C.sqlite3_malloc(size * C.int(n)))
	pages := unsafePointerToSlice(pList, n)
	for i := range pages {
		page := &(pages[i])
		page.pBuf = unsafe.Pointer(C.sqlite3_malloc(C.int(pageSize)))
	}
	return pages
}

// DestroyReplicationPages deallocates the memory of the given
// ReplicationPage slice.
func DestroyReplicationPages(pages []ReplicationPage) {
	for i := range pages {
		page := &(pages[i])
		C.sqlite3_free(page.pBuf)
	}
	C.sqlite3_free(unsafe.Pointer(&pages[0]))
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

	// The cgo specification forbids to store Go pointers in C memory (see
	// https://golang.org/cmd/cgo/#hdr-Passing_pointers), so we perform
	// the memory allocation of the sqlite3_replication_methods C structure
	// using sqlite3_malloc, and set its pArg to be itself. Then, we use
	// a registry in Go memory for mapping the sqlite3_replication_methods pointers
	// created here to the associated SQLiteConn and Methods objects.
	size := C.int(unsafe.Sizeof(C.sqlite3_replication_methods{}))
	pMethods := (*C.sqlite3_replication_methods)(unsafe.Pointer(C.sqlite3_malloc(size)))

	// Set the callbacks
	pMethods.pArg = unsafe.Pointer(db)
	pMethods.xBegin = (*[0]byte)(unsafe.Pointer(C.replicationBegin))
	pMethods.xWalFrames = (*[0]byte)(unsafe.Pointer(C.replicationWalFrames))
	pMethods.xUndo = (*[0]byte)(unsafe.Pointer(C.replicationUndo))
	pMethods.xEnd = (*[0]byte)(unsafe.Pointer(C.replicationEnd))
	pMethods.xCheckpoint = (*[0]byte)(unsafe.Pointer(C.replicationCheckpoint))

	if rc := C.sqlite3_replication_leader(db, C.CString("main"), pMethods); rc != C.SQLITE_OK {
		C.sqlite3_free(unsafe.Pointer(pMethods))
		return Error{Code: ErrNo(int(rc))}
	}

	registerMethodsInstance(conn, pMethods, methods)

	return nil
}

// ReplicationFollower switches the given sqlite connection to
// follower replication mode. In this mode no regular operation is
// possible, and the connection should be driven with the
// ReplicationBegin, ReplicationWalFrames, ReplicationCommit and
// ReplicationRollback APIs.
func ReplicationFollower(conn *SQLiteConn) error {
	db := conn.db
	if rc := C.sqlite3_replication_follower(db, C.CString("main")); rc != C.SQLITE_OK {
		return Error{Code: ErrNo(int(rc))}
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

	if rc := C.sqlite3_replication_none(db, C.CString("main")); rc != C.SQLITE_OK {
		return nil, Error{Code: ErrNo(rc)}
	}

	C.sqlite3_free(unsafe.Pointer(instance.pMethods))
	unregisterMethodsInstance(conn)
	return instance.methods, nil
}

// ReplicationMode returns the current replication mode of the given
// transaction.
func ReplicationMode(conn *SQLiteConn) (Replication, error) {
	db := conn.db
	mode := new(int)

	// Convert to C pointer
	eMode := (*C.int)(unsafe.Pointer(mode))
	if rc := C.sqlite3_replication_mode(db, C.CString("main"), eMode); rc != C.SQLITE_OK {
		return 0, Error{Code: ErrNo(int(rc))}
	}
	return Replication(*mode), nil
}

// ReplicationBegin starts a new write transaction in the given sqlite
// connection. This should be called against a "follower" connection,
// meant to replicate the "leader" one.
func ReplicationBegin(conn *SQLiteConn) error {
	if rc := C.sqlite3_replication_begin(conn.db, C.CString("main")); rc != C.SQLITE_OK {
		return Error{Code: ErrNo(rc)}
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
	zDb := C.CString("main")
	szPage := C.int(params.PageSize)
	nList := C.int(len(params.Pages))
	nTruncate := C.uint(params.Truncate)
	pList := (*C.sqlite3_replication_page)(unsafe.Pointer(&params.Pages[0]))
	isCommit := C.int(params.IsCommit)
	syncFlags := C.int(params.SyncFlags)

	if rc := C.sqlite3_replication_wal_frames(db, zDb, szPage, nList, pList, nTruncate, isCommit, syncFlags); rc != C.SQLITE_OK {
		return Error{Code: ErrNo(rc)}
	}
	return nil
}

// ReplicationUndo rollbacks a write transaction in the given sqlite
// connection. This should be called with a "follower" connection,
// meant to replicate the "leader" one.
func ReplicationUndo(conn *SQLiteConn) error {
	if rc := C.sqlite3_replication_undo(conn.db, C.CString("main")); rc != C.SQLITE_OK {
		return Error{Code: ErrNo(rc)}
	}
	return nil
}

// ReplicationEnd finishes a write transaction in the given sqlite
// connection. This should be called with a "follower" connection,
// meant to replicate the "leader" one.
func ReplicationEnd(conn *SQLiteConn) error {
	if rc := C.sqlite3_replication_end(conn.db, C.CString("main")); rc != C.SQLITE_OK {
		return Error{Code: ErrNo(rc)}
	}
	return nil
}

// ReplicationCheckpoint checkpoints the current WAL.
func ReplicationCheckpoint(conn *SQLiteConn, mode WalCheckpointMode, log *int, ckpt *int) error {
	// Convert to C types
	db := conn.db
	zDb := C.CString("main")
	eMode := C.int(mode)
	pnLog := (*C.int)(unsafe.Pointer(log))
	pnCkpt := (*C.int)(unsafe.Pointer(ckpt))

	if rc := C.sqlite3_replication_checkpoint(db, zDb, eMode, pnLog, pnCkpt); rc != C.SQLITE_OK {
		return Error{Code: ErrNo(rc)}
	}
	return nil
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

	// This seems the easiest way to convert a C pointer to a slice without
	// copying the data https://github.com/golang/go/issues/13656. Note that
	// the size of the array is 2^30, which should be virtually enough for
	// any situation.
	pages := unsafePointerToSlice(unsafe.Pointer(pList), int(nList))

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
	pMethods *C.sqlite3_replication_methods
	conn     *SQLiteConn
	methods  ReplicationMethods
}

// A registry for tracking instances of C.sqlite3_replication_methods
// allocated by this module.
//
// Each key is a pointer to a C.sqlite3 db object, and each associated
// value a replicationMethodsInstance holding a pointer to the C callbacks
// were registered against that db, and the Methods object that those
// callbacks will be dispatched to.
var replicationMethodsRegistry = make(map[*C.sqlite3]*replicationMethodsInstance)

// Serialize access to the methods registry
var replicationMethodsMutex sync.RWMutex

// Register a new replication for the given connection and methods.
func registerMethodsInstance(conn *SQLiteConn, pMethods *C.sqlite3_replication_methods, methods ReplicationMethods) {
	replicationMethodsMutex.Lock()
	defer replicationMethodsMutex.Unlock()

	replicationMethodsRegistry[conn.db] = &replicationMethodsInstance{
		pMethods: pMethods,
		conn:     conn,
		methods:  methods,
	}
}

// Unregister a previously registered replication
func unregisterMethodsInstance(conn *SQLiteConn) {
	replicationMethodsMutex.Lock()
	defer replicationMethodsMutex.Unlock()

	delete(replicationMethodsRegistry, conn.db)
}

// Find the replication methods instance for the given database.
func findMethodsInstance(db *C.sqlite3) *replicationMethodsInstance {
	replicationMethodsMutex.RLock()
	defer replicationMethodsMutex.RUnlock()

	instance, _ := replicationMethodsRegistry[db]
	return instance
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
