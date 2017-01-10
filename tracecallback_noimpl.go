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

// RegisterAggregator register the aggregator.
func (c *SQLiteConn) RegisterAggregator(name string, impl interface{}, pure bool) error {
	return errors.New("This feature is not implemented")
}

func (c *SQLiteConn) SetTrace(requested *TraceConfig) error {
	return errors.New("This feature is not implemented")
}
