// +build !trace

package sqlite3

import "errors"

// Trace... constants identify the possible events causing callback invocation.
// Values are same as the corresponding SQLite Trace Event Codes.
const (
	TraceStmt    = uint32(0x01)
	TraceProfile = uint32(0x02)
	TraceRow     = uint32(0x04)
	TraceClose   = uint32(0x08)
)

type TraceInfo struct {
	// Pack together the shorter fields, to keep the struct smaller.
	// On a 64-bit machine there would be padding
	// between EventCode and ConnHandle; having AutoCommit here is "free":
	EventCode  uint32
	AutoCommit bool
	ConnHandle uintptr

	// Usually filled, unless EventCode = TraceClose = SQLITE_TRACE_CLOSE:
	// identifier for a prepared statement:
	StmtHandle uintptr

	// Two strings filled when EventCode = TraceStmt = SQLITE_TRACE_STMT:
	// (1) either the unexpanded SQL text of the prepared statement, or
	//     an SQL comment that indicates the invocation of a trigger;
	// (2) expanded SQL, if requested and if (1) is not an SQL comment.
	StmtOrTrigger string
	ExpandedSQL   string // only if requested (TraceConfig.WantExpandedSQL = true)

	// filled when EventCode = TraceProfile = SQLITE_TRACE_PROFILE:
	// estimated number of nanoseconds that the prepared statement took to run:
	RunTimeNanosec int64

	DBError Error
}

type TraceUserCallback func(TraceInfo) int

type TraceConfig struct {
	Callback        TraceUserCallback
	EventMask       uint
	WantExpandedSQL bool
}

// RegisterAggregator register the aggregator.
func (c *SQLiteConn) RegisterAggregator(name string, impl interface{}, pure bool) error {
	return errors.New("This feature is not implemented")
}

func (c *SQLiteConn) SetTrace(requested *TraceConfig) error {
	return errors.New("This feature is not implemented")
}
