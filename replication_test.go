package sqlite3_test

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/CanonicalLtd/go-sqlite3"
)

func TestReplicationLeader_CannotSetNonWalJournal(t *testing.T) {
	conn, cleanup := newFileSQLiteConn()
	defer cleanup()

	err := sqlite3.ReplicationLeader(conn, sqlite3.PassthroughReplicationMethods())

	if err == nil {
		t.Fatal("expected error when trying to set replication leader mode on non-WAL database")
	}
	want := sqlite3.ErrorString(sqlite3.ErrMisuse)
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestReplicationLeader_CannotSetTwice(t *testing.T) {
	methods := sqlite3.PassthroughReplicationMethods()
	conn, cleanup := newLeaderSQLiteConn(methods)
	defer cleanup()

	err := sqlite3.ReplicationLeader(conn, methods)

	if err == nil {
		t.Fatal("expected error when trying to set replication leader mode twice")
	}
	const want = "leader replication already enabled for this connection"
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestReplicationNone_CannotSetWithNonLeader(t *testing.T) {
	conn, cleanup := newFileSQLiteConn()
	defer cleanup()

	_, err := sqlite3.ReplicationNone(conn)

	if err == nil {
		t.Fatal("expected error when trying to set replication none on non-leader connection")
	}
	const want = "leader replication is not enabled for this connection"
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestReplicationFollower_CannotSetWithNonWalJournal(t *testing.T) {
	conn, cleanup := newFileSQLiteConn()
	defer cleanup()

	err := sqlite3.ReplicationFollower(conn)

	if err == nil {
		t.Fatal("expected error when trying to set replication follower mode on non-WAL database")
	}
	want := sqlite3.ErrorString(sqlite3.ErrMisuse)
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestReplicationFollower_DisableRegularOperations(t *testing.T) {
	conn, cleanup := newFollowerSQLiteConn()
	defer cleanup()

	_, err := conn.Exec("SELECT * FROM sqlite_master", nil)

	if err == nil {
		t.Fatal("expected error when trying to perform a query in replication follower mode")
	}
	const want = "database is in follower replication mode: main"
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestReplicationNone_CannotUseIfReplicationModeIsFollower(t *testing.T) {
	conn, cleanup := newFollowerSQLiteConn()
	defer cleanup()

	defer sqlite3.ReplicationRegisterMethodsInstance(conn)()
	_, err := sqlite3.ReplicationNone(conn)

	if err == nil {
		t.Fatal("expected error when trying to use ReplicationNone with replication mode follower")
	}
	want := sqlite3.ErrorString(sqlite3.ErrMisuse)
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestReplicationMode(t *testing.T) {
	var mode sqlite3.Replication
	hook := func(conn *sqlite3.SQLiteConn) sqlite3.ErrNo {
		var err error
		mode, err = sqlite3.ReplicationMode(conn)
		if err != nil {
			t.Fatal(err)
		}
		return sqlite3.ErrInternal
	}
	methods := &hookReplicationMethods{hook: hook}
	conn, cleanup := newLeaderSQLiteConn(methods)
	defer cleanup()

	if _, err := conn.Exec("CREATE TABLE foo (int N)", nil); err == nil {
		t.Fatal("expected create table to trigger failing hook")
	}

	if mode != sqlite3.ReplicationModeLeader {
		t.Errorf("got replication mode %d instead of leader", mode)
	}
}

func TestReplicationBegin_MethodsInstanceNotRegistered(t *testing.T) {
	const want = "replication hooks not found"
	defer func() {
		got := recover()
		if got != want {
			t.Errorf("expected\n%q\ngot\n%q", want, got)
		}
	}()
	sqlite3.ReplicationBeginHook(unsafe.Pointer(nil))
}

/*
func TestReplication_CannotUseIfReplicationModeIsNone(t *testing.T) {
	// Wrapper around ReplicationWalFrames providing test parameters
	replicationWalFrames := func(conn *sqlite3.SQLiteConn) error {
		params := &sqlite3.ReplicationWalFramesParams{
			Pages: sqlite3.NewReplicationPages(2, 4096),
		}
		return sqlite3.ReplicationWalFrames(conn, params)
	}

	cases := []struct {
		name   string
		method func(*sqlite3.SQLiteConn) error
	}{
		{"ReplicationBegin", sqlite3.ReplicationBegin},
		{"ReplicationWalFrames", replicationWalFrames},
		{"ReplicationUndo", sqlite3.ReplicationEnd},
		{"ReplicationEnd", sqlite3.ReplicationEnd},
	}

	for _, c := range cases {
		subtest.Run(t, c.name, func(t *testing.T) {
			conn, cleanup := newFileSQLiteConn()
			defer cleanup()

			if err := sqlite3.JournalModePragma(conn, sqlite3.JournalWal); err != nil {
				t.Fatalf("failed to set WAL journal mode: %v", err)
			}

			err := c.method(conn)
			if err == nil {
				t.Fatalf("expected error when trying to use %s with replication mode none", c.name)
			}
			want := sqlite3.ErrorString(sqlite3.ErrMisuse)
			got := err.Error()
			if got != want {
				t.Errorf("expected\n%q\ngot\n%q", want, got)
			}
		})
	}
}
*/

func TestReplicationPages(t *testing.T) {
	pages := sqlite3.NewReplicationPages(2, 4096)
	if len(pages) != 2 {
		t.Fatalf("Got %d pages instead of 2", len(pages))
	}
	for i := range pages {
		page := pages[i]
		if page.Data() == nil {
			t.Errorf("The data buffer for page %d is not allocated", i)
		}
		page.Fill([]byte("hello"), 1, uint32(i))
		if page.Data() == nil {
			t.Errorf("The data buffer for page %d is NULL", i)
		}
		if page.Flags() != 1 {
			t.Errorf("Failed to fill flags for page %d", i)
		}
		if page.Number() != uint32(i) {
			t.Errorf("Failed to fill number for page %d", i)
		}
	}
}

func TestReplicationMethods_Commit(t *testing.T) {
	cluster, cleanup := newReplicationCluster()
	defer cleanup()

	mustExec(cluster.Leader, "BEGIN; CREATE TABLE test (n INT); INSERT INTO test VALUES(1); COMMIT", nil)

	rows := mustQuery(cluster.Observer, "SELECT * FROM test", nil)
	defer rows.Close()

	if err := rows.Next(nil); err != nil {
		t.Fatal(err)
	}
}

func TestReplicationMethods_Rollback(t *testing.T) {
	cluster, cleanup := newReplicationCluster()
	defer cleanup()

	mustExec(cluster.Leader, "CREATE TABLE test (n INT)", nil)
	mustExec(cluster.Leader, "BEGIN; INSERT INTO test VALUES(1); ROLLBACK", nil)

	rows := mustQuery(cluster.Observer, "SELECT * FROM test", nil)
	defer rows.Close()

	if err := rows.Next(nil); err == nil {
		t.Fatal("follower database was not rolled back")
	}
}

func TestReplicationCheckpoint(t *testing.T) {
	cluster, cleanup := newReplicationCluster()
	defer cleanup()

	sqlite3.ReplicationAutoCheckpoint(cluster.Leader, 1)

	mustExec(cluster.Leader, "CREATE TABLE test (n INT)", nil)
	if sqlite3.WalSize(cluster.Leader) != 0 {
		t.Fatal("leader WAL file was not truncated")
	}
	if sqlite3.WalSize(cluster.Follower) != 0 {
		t.Fatal("follower WAL file was not truncated")
	}
}

func TestReplicationCheckpoint_Error(t *testing.T) {
	conn1, cleanup1 := newLeaderSQLiteConn(sqlite3.PassthroughReplicationMethods())
	defer cleanup1()

	conn2 := newSQLiteConn(sqlite3.DatabaseFilename(conn1))

	if _, err := conn1.Exec("CREATE TABLE test (n INT)", nil); err != nil {
		t.Fatal(err)
	}

	sqlite3.ReplicationAutoCheckpoint(conn1, 1)
	if err := sqlite3.BusyTimeoutPragma(conn1, 1); err != nil {
		t.Fatal(err)
	}

	if _, err := conn1.Exec("BEGIN", nil); err != nil {
		t.Fatal(err)
	}

	if _, err := conn2.Exec("BEGIN", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := conn2.Exec("SELECT * FROM test", nil); err != nil {
		t.Fatal(err)
	}

	if _, err := conn1.Exec("INSERT INTO test VALUES(1)", nil); err != nil {
		t.Fatal(err)
	}

	// The PassthroughReplicationMethods.Checkpoint callback should panic
	// on error.
	defer func() {
		err, ok := recover().(sqlite3.Error)
		if !ok {
			t.Fatal("expected checkpoint to panic with a sqlite3.Error")
		}
		if err.Code != sqlite3.ErrBusy {
			t.Fatalf("expected SQLite ErrBusy, got %d", err.Code)
		}
	}()
	mustExec(conn1, "COMMIT", nil)
}

func TestPassthroughReplicationMethods(t *testing.T) {
	methods := sqlite3.PassthroughReplicationMethods()
	conn, cleanup := newLeaderSQLiteConn(methods)
	defer cleanup()

	// Set auto-checkpoint to 1 so we trigger the checkpoint hook too.
	sqlite3.ReplicationAutoCheckpoint(conn, 1)

	mustExec(conn, "BEGIN; CREATE TABLE test (n INT); COMMIT", nil)
	mustQuery(conn, "BEGIN; SELECT * FROM test; COMMIT;", nil)
	mustExec(conn, "BEGIN; INSERT INTO test VALUES(1); ROLLBACK", nil)
}

/*
func TestPassthroughReplicationMethods_Panic(t *testing.T) {
	methods := sqlite3.PassthroughReplicationMethods()
	cases := map[string]func(*sqlite3.SQLiteConn){
		"begin": func(c *sqlite3.SQLiteConn) {
			methods.Begin(c)
		},
		"wal frames": func(c *sqlite3.SQLiteConn) {
			params := &sqlite3.ReplicationWalFramesParams{
				Pages: sqlite3.NewReplicationPages(2, 4096),
			}
			methods.WalFrames(c, params)
		},
		"undo": func(c *sqlite3.SQLiteConn) {
			methods.Undo(c)
		},
		"end": func(c *sqlite3.SQLiteConn) {
			methods.End(c)
		},
		"checkpoint": func(c *sqlite3.SQLiteConn) {
			methods.Checkpoint(c, sqlite3.WalCheckpointPassive, nil, nil)
		},
	}

	for name, method := range cases {
		subtest.Run(t, name, func(t *testing.T) {
			conn := newMemorySQLiteConn()
			want := sqlite3.ErrorString(sqlite3.ErrMisuse)
			defer func() {
				err, ok := recover().(error)
				if !ok {
					t.Error("recover() return value is not an error")
				}
				got := err.Error()
				if got != want {
					t.Errorf("expected panic\n%q\ngot\n%q", want, got)
				}
			}()
			method(conn)
		})
	}

}
*/

// A cluster of replicated connections composed by one leader, one
// follower and one observer connection connected to the same database
// as the follower.
type replicationCluster struct {
	Leader   *sqlite3.SQLiteConn
	Follower *sqlite3.SQLiteConn
	Observer *sqlite3.SQLiteConn
}

// Return a SQLiteConn opened against a temporary database filename,
// set to WAL journal mode and configured for follower replication.
func newFollowerSQLiteConn() (*sqlite3.SQLiteConn, func()) {
	conn, cleanup := newFileSQLiteConn()

	if err := sqlite3.JournalModePragma(conn, sqlite3.JournalWal); err != nil {
		panic(fmt.Sprintf("failed to set WAL journal mode: %v", err))
	}

	if err := sqlite3.DatabaseNoCheckpointOnClose(conn); err != nil {
		panic(fmt.Sprintf("failed to set disable checkpoint on close: %v", err))
	}

	if err := sqlite3.ReplicationFollower(conn); err != nil {
		panic(fmt.Sprintf("failed to set follower replication mode: %v", err))
	}
	return conn, cleanup
}

// Return a SQLiteConn opened against a temporary database filename,
// set to WAL journal mode and configured for leader replication.
func newLeaderSQLiteConn(methods sqlite3.ReplicationMethods) (*sqlite3.SQLiteConn, func()) {
	conn, connCleanup := newFileSQLiteConn()

	if err := sqlite3.JournalModePragma(conn, sqlite3.JournalWal); err != nil {
		panic(fmt.Sprintf("failed to set WAL journal mode: %v", err))
	}

	if err := sqlite3.ReplicationLeader(conn, methods); err != nil {
		panic(fmt.Sprintf("failed to set leader replication mode: %v", err))
	}
	cleanup := func() {
		if _, err := sqlite3.ReplicationNone(conn); err != nil {
			panic(fmt.Sprintf("failed to set replication mode to none: %v", err))
		}
		connCleanup()
	}
	return conn, cleanup
}

func newReplicationCluster() (*replicationCluster, func()) {
	follower, followerCleanup := newFollowerSQLiteConn()
	methods := &directReplicationMethods{followers: []*sqlite3.SQLiteConn{follower}}
	leader, leaderCleanup := newLeaderSQLiteConn(methods)
	observer := newSQLiteConn(sqlite3.DatabaseFilename(follower))

	cluster := &replicationCluster{
		Follower: follower,
		Leader:   leader,
		Observer: observer,
	}

	cleanup := func() {
		leaderCleanup()
		followerCleanup()
	}

	return cluster, cleanup
}

// ReplicationMethods implementation that replicates WAL commands directly
// across the given follower connections.
type directReplicationMethods struct {
	followers []*sqlite3.SQLiteConn
}

func (m *directReplicationMethods) Begin(conn *sqlite3.SQLiteConn) sqlite3.ErrNo {
	conns := append(m.followers, conn)
	for i, conn := range conns {
		if err := sqlite3.ReplicationBegin(conn); err != nil {
			panic(fmt.Sprintf("begin failed on conn %d: %v", i, err))
		}
	}
	return 0
}

func (m *directReplicationMethods) WalFrames(conn *sqlite3.SQLiteConn, params *sqlite3.ReplicationWalFramesParams) sqlite3.ErrNo {
	conns := append(m.followers, conn)
	for i, conn := range conns {
		if err := sqlite3.ReplicationWalFrames(conn, params); err != nil {
			panic(fmt.Sprintf("wal frames failed on conn %d: %v", i, err))
		}
	}
	return 0
}

func (m *directReplicationMethods) Undo(conn *sqlite3.SQLiteConn) sqlite3.ErrNo {
	conns := append(m.followers, conn)
	for i, conn := range conns {
		if err := sqlite3.ReplicationUndo(conn); err != nil {
			panic(fmt.Sprintf("undo failed on conn %d: %v", i, err))
		}
	}
	return 0
}

func (m *directReplicationMethods) End(conn *sqlite3.SQLiteConn) sqlite3.ErrNo {
	conns := append(m.followers, conn)
	for i, conn := range conns {
		if err := sqlite3.ReplicationEnd(conn); err != nil {
			panic(fmt.Sprintf("end failed on conn %d: %v", i, err))
		}
	}
	return 0
}

func (m *directReplicationMethods) Checkpoint(conn *sqlite3.SQLiteConn, mode sqlite3.WalCheckpointMode, log *int, ckpt *int) sqlite3.ErrNo {
	conns := append(m.followers, conn)
	for i, conn := range conns {
		if err := sqlite3.ReplicationCheckpoint(conn, mode, log, ckpt); err != nil {
			panic(fmt.Sprintf("checkpoint failed on conn %d: %v", i, err))
		}
	}
	return 0
}

// ReplicationMethods implementation that defers execution to the given hook.
type hookReplicationMethods struct {
	hook func(*sqlite3.SQLiteConn) sqlite3.ErrNo
}

func (m *hookReplicationMethods) Begin(conn *sqlite3.SQLiteConn) sqlite3.ErrNo {
	return m.hook(conn)
}

func (m *hookReplicationMethods) WalFrames(conn *sqlite3.SQLiteConn, params *sqlite3.ReplicationWalFramesParams) sqlite3.ErrNo {
	return m.hook(conn)
}

func (m *hookReplicationMethods) Undo(conn *sqlite3.SQLiteConn) sqlite3.ErrNo {
	return m.hook(conn)
}

func (m *hookReplicationMethods) End(conn *sqlite3.SQLiteConn) sqlite3.ErrNo {
	return m.hook(conn)
}

func (m *hookReplicationMethods) Checkpoint(conn *sqlite3.SQLiteConn, mode sqlite3.WalCheckpointMode, log *int, ckpt *int) sqlite3.ErrNo {
	return m.hook(conn)
}
