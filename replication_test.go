package sqlite3

import (
	"fmt"
	"os"
	"testing"
)

func TestReplicationPages(t *testing.T) {
	pages := NewReplicationPages(2, 4096)
	if len(pages) != 2 {
		t.Fatalf("Got %d pages instead of 2", len(pages))
	}
	for i, page := range pages {
		if page.Data() == nil {
			t.Errorf("The data buffer for page %d is not allocated", i)
		}
		if n := len(page.Data()); n != 4096 {
			t.Errorf("The data buffer for page %d has unexpected lenght %d", i, n)
		}
		page.Fill([]byte("hello"), 1, uint32(i))
		if string(page.Data()) != "hello" {
			t.Errorf("The data buffer for page %d does not match the given bytes", i)
		}
		if page.Flags() != 1 {
			t.Errorf("Failed to fill flags for page %d", i)
		}
		if page.Number() != uint32(i) {
			t.Errorf("Failed to fill number for page %d", i)
		}
	}
}

func TestReplicationModesErrors(t *testing.T) {
	cases := []struct {
		name string                                     // Name of the test
		f    func(t *testing.T, conn *SQLiteConn) error // Scenario leading to an error
	}{
		{
			"connection not in WAL mode: follower",
			func(t *testing.T, conn *SQLiteConn) error {
				return conn.ReplicationLeader(NoopReplicationMethods())
			},
		},
		{
			"connection not in WAL mode: leader",
			func(t *testing.T, conn *SQLiteConn) error {
				return conn.ReplicationFollower()
			},
		},
		{
			"cannot set twice: leader",
			func(t *testing.T, conn *SQLiteConn) error {
				pragmaWAL(t, conn)
				err := conn.ReplicationLeader(NoopReplicationMethods())
				if err != nil {
					t.Fatal("failed to set leader replication:", err)
				}
				return conn.ReplicationLeader(NoopReplicationMethods())
			},
		},
		{
			"cannot set twice: follower",
			func(t *testing.T, conn *SQLiteConn) error {
				pragmaWAL(t, conn)
				err := conn.ReplicationFollower()
				if err != nil {
					t.Fatal("failed to set follower replication:", err)
				}
				return conn.ReplicationFollower()
			},
		},
		{
			"cannot switch from leader to follower",
			func(t *testing.T, conn *SQLiteConn) error {
				pragmaWAL(t, conn)
				err := conn.ReplicationLeader(NoopReplicationMethods())
				if err != nil {
					t.Fatal("failed to set leader replication:", err)
				}
				return conn.ReplicationFollower()
			},
		},
		{
			"cannot switch from follower to leader",
			func(t *testing.T, conn *SQLiteConn) error {
				pragmaWAL(t, conn)
				err := conn.ReplicationFollower()
				if err != nil {
					t.Fatal("failed to set follower replication:", err)
				}
				return conn.ReplicationLeader(NoopReplicationMethods())
			},
		},
		{
			"cannot run queries as follower",
			func(t *testing.T, conn *SQLiteConn) error {
				pragmaWAL(t, conn)
				err := conn.ReplicationFollower()
				if err != nil {
					t.Fatal("failed to set follower replication:", err)
				}
				_, err = conn.Query("SELECT 1", nil)
				return err
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tempFilename := TempFilename(t)
			defer os.Remove(tempFilename)

			driver := &SQLiteDriver{}
			conn, err := driver.Open(tempFilename)
			if err != nil {
				t.Fatalf("can't open connection to %s: %v", tempFilename, err)
			}
			conni := conn.(*SQLiteConn)
			defer conni.Close()

			err = c.f(t, conni)
			if err == nil {
				t.Fatal("no error was returned")
			}
			erri, ok := err.(Error)
			if !ok {
				t.Fatalf("returned error %#v is not of type Error", erri)
			}
			if erri.Code != ErrError {
				t.Errorf("expected error code %d, got %d", ErrError, erri.Code)
			}

		})
	}
}

func TestReplicationMethods(t *testing.T) {
	conns := make([]*SQLiteConn, 2) // Index 0 is the leader and index 1 is the follower

	// Open the connections.
	driver := &SQLiteDriver{}
	for i := range conns {
		tempFilename := TempFilename(t)
		defer os.Remove(tempFilename)
		conn, err := driver.Open(tempFilename)
		if err != nil {
			t.Fatalf("can't open connection to %s: %v", tempFilename, err)
		}
		defer conn.Close()

		conni := conn.(*SQLiteConn)
		pragmaWAL(t, conni)
		conns[i] = conni
	}
	leader := conns[0]
	follower := conns[1]

	// Set leader replication on conn 0.
	methods := &directReplicationMethods{
		follower: follower,
	}
	if err := leader.ReplicationLeader(methods); err != nil {
		t.Fatal("failed to switch to leader replication:", err)
	}

	// Set follower replication on conn 1.
	if err := follower.ReplicationFollower(); err != nil {
		t.Fatal("failed to switch to follower replication:", err)
	}

	// Create a table on the leader.
	if _, err := leader.Exec("CREATE TABLE a (n INT)", nil); err != nil {
		t.Fatal("failed to execute query on leader:", err)
	}

	// Rollback a transaction on the leader.
	if _, err := leader.Exec("BEGIN; CREATE TABLE b (n INT); ROLLBACK", nil); err != nil {
		t.Fatal("failed to rollback query on leader:", err)
	}

	// Check that the follower has replicated the commit but not the rollback.
	if err := follower.ReplicationNone(); err != nil {
		t.Fatal("failed to turn off follower replication:", err)
	}
	if _, err := follower.Query("SELECT 1", nil); err != nil {
		t.Fatal("failed to execute query on follower:", err)
	}
	if _, err := follower.Query("SELECT n FROM a", nil); err != nil {
		t.Fatal("failed to execute query on follower:", err)
	}
	if _, err := follower.Query("SELECT n FROM b", nil); err == nil {
		t.Fatal("expected error when querying rolled back table:", err)
	}
}

// ReplicationMethods implementation that replicates WAL commands directly
// to the given follower.
type directReplicationMethods struct {
	follower *SQLiteConn
	writing  bool
}

func (m *directReplicationMethods) Begin(conn *SQLiteConn) ErrNo {
	return 0
}

func (m *directReplicationMethods) Abort(conn *SQLiteConn) ErrNo {
	return 0
}

func (m *directReplicationMethods) Frames(conn *SQLiteConn, params *ReplicationFramesParams) ErrNo {
	begin := false
	if !m.writing {
		begin = true
		m.writing = true
	}
	if err := ReplicationFrames(m.follower, begin, params); err != nil {
		panic(fmt.Sprintf("frames failed: %v", err))
	}
	if params.IsCommit == 1 {
		m.writing = false
	}
	return 0
}

func (m *directReplicationMethods) Undo(conn *SQLiteConn) ErrNo {
	if m.writing {
		if err := ReplicationUndo(m.follower); err != nil {
			panic(fmt.Sprintf("undo failed: %v", err))
		}
	}
	return 0
}

func (m *directReplicationMethods) End(conn *SQLiteConn) ErrNo {
	return 0
}

// An xBegin error never triggers an xUndo callback and SQLite takes care of
// releasing the WAL write lock, if acquired.
func TestReplicationMethods_BeginError(t *testing.T) {
	// Open the leader connection.
	cases := []struct {
		errno ErrNoExtended
		lock  bool
	}{
		{ErrConstraintCheck, false},
		{ErrConstraintCheck, true},
		{ErrCorruptVTab, true},
		{ErrCorruptVTab, false},
		{ErrIoErrNotLeader, true},
		{ErrIoErrNotLeader, false},
		{ErrIoErrLeadershipLost, true},
		{ErrIoErrLeadershipLost, false},
		{ErrIoErrRead, true},
		{ErrIoErrRead, false},
		{ErrIoErrWrite, true},
		{ErrIoErrWrite, false},
	}

	for _, c := range cases {
		name := fmt.Sprintf("%s-%v", c.errno, c.lock)
		t.Run(name, func(t *testing.T) {
			// Create a leader connection with the appropriate
			// replication methods.
			driver := &SQLiteDriver{}
			tempFilename := TempFilename(t)
			defer os.Remove(tempFilename)

			conni, err := driver.Open(tempFilename)
			if err != nil {
				t.Fatalf("can't open connection to %s: %v", tempFilename, err)
			}
			defer conni.Close()

			conn := conni.(*SQLiteConn)
			pragmaWAL(t, conn)

			// Set leader replication on conn 0.
			methods := &failingReplicationMethods{
				conn:  conn,
				hook:  "begin",
				lock:  c.lock,
				errno: c.errno,
			}
			if err := conn.ReplicationLeader(methods); err != nil {
				t.Fatal("failed to switch to leader replication:", err)
			}

			// Execute a query that should error and be rolled back.
			tx, err := conn.Begin()
			if err != nil {
				t.Fatal("failed to begin transaction", err)
			}
			_, err = conn.Exec("CREATE TABLE test (n INT)", nil)
			erri, ok := err.(Error)
			if !ok {
				t.Fatalf("returned error %#v is not of type Error", erri)
			}
			if erri.ExtendedCode != c.errno {
				t.Errorf("expected error code %d, got %d", c.errno, erri.ExtendedCode)
			}
			err = tx.Rollback()

			// ErrIo errors will also fail to rollback, while other
			// errors are fine.
			if erri.Code == ErrIoErr {
				if err == nil {
					t.Fatal("expected rollback error")
				}
				if err.Error() != "cannot rollback - no transaction is active" {
					t.Fatal("expected different rollback error")
				}
			} else {
				if err != nil {
					t.Fatal("failed to rollback", err)
				}
			}

			// Execute a second query with no error.
			methods.hook = ""
			tx, err = conn.Begin()
			if err != nil {
				t.Fatal("failed to begin transaction", err)
			}
			_, err = conn.Exec("CREATE TABLE test (n INT)", nil)
			if err != nil {
				t.Fatal("failed to execute query", err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal("failed to commit transaction", err)
			}
		})
	}
}

// An xFrames error triggers the xUndo callback.
func TestReplicationMethods_FramesError(t *testing.T) {
	// Create a leader connection with the appropriate
	// replication methods.
	driver := &SQLiteDriver{}
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	conni, err := driver.Open(tempFilename)
	if err != nil {
		t.Fatalf("can't open connection to %s: %v", tempFilename, err)
	}
	defer conni.Close()

	conn := conni.(*SQLiteConn)
	pragmaWAL(t, conn)

	// Set leader replication on conn 0.
	methods := &failingReplicationMethods{
		conn:  conn,
		hook:  "frames",
		errno: ErrIoErrNotLeader,
	}
	if err := conn.ReplicationLeader(methods); err != nil {
		t.Fatal("failed to switch to leader replication:", err)
	}
	_, err = conn.Exec("CREATE TABLE test (n INT)", nil)
	erri, ok := err.(Error)
	if !ok {
		t.Fatalf("returned error %#v is not of type Error", erri)
	}
	if erri.ExtendedCode != ErrIoErrNotLeader {
		t.Errorf("expected error code %d, got %d", ErrIoErrNotLeader, erri.ExtendedCode)
	}
	if n := len(methods.fired); n != 4 {
		t.Fatalf("expected 4 hooks to be fired, instead of %d", n)
	}
	hooks := []string{"begin", "frames", "undo", "end"}
	for i := range methods.fired {
		if hook := methods.fired[i]; hook != hooks[i] {
			t.Errorf("expected hook %s to be fired, instead of %s", hooks[i], hook)
		}
	}
}

// If an xUndo hook fails, the ROLLBACK query still succeeds.
func TestReplicationMethods_UndoError(t *testing.T) {
	// Create a leader connection with the appropriate
	// replication methods.
	driver := &SQLiteDriver{}
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	conni, err := driver.Open(tempFilename)
	if err != nil {
		t.Fatalf("can't open connection to %s: %v", tempFilename, err)
	}
	defer conni.Close()

	conn := conni.(*SQLiteConn)
	pragmaWAL(t, conn)

	// Set leader replication on conn 0.
	methods := &failingReplicationMethods{
		conn:  conn,
		hook:  "undo",
		errno: ErrIoErrNotLeader,
	}
	if err := conn.ReplicationLeader(methods); err != nil {
		t.Fatal("failed to switch to leader replication:", err)
	}
	_, err = conn.Exec("BEGIN; CREATE TABLE test (n INT); ROLLBACK", nil)
	if err != nil {
		t.Fatal("rollback failed", err)
	}
	if n := len(methods.fired); n != 3 {
		t.Fatalf("expected 3 hooks to be fired, instead of %d", n)
	}
	hooks := []string{"begin", "undo", "end"}
	for i := range methods.fired {
		if hook := methods.fired[i]; hook != hooks[i] {
			t.Errorf("expected hook %s to be fired, instead of %s", hooks[i], hook)
		}
	}
}

// ReplicationMethods implementation that fails in a programmable way.
type failingReplicationMethods struct {
	conn  *SQLiteConn   // Leader connection
	lock  bool          // Whether to acquire the WAL write lock before failing
	hook  string        // Name of the hook that should fail
	errno ErrNoExtended // Error to be returned by the hook
	fired []string      // Hooks that were fired
}

func (m *failingReplicationMethods) Begin(conn *SQLiteConn) ErrNo {
	m.fired = append(m.fired, "begin")

	if m.hook == "begin" {
		return ErrNo(m.errno)
	}

	return 0
}

func (m *failingReplicationMethods) Abort(conn *SQLiteConn) ErrNo {
	return 0
}

func (m *failingReplicationMethods) Frames(conn *SQLiteConn, params *ReplicationFramesParams) ErrNo {
	m.fired = append(m.fired, "frames")

	if m.hook == "begin" {
		panic("frames hook should not be reached")
	}
	if m.hook == "frames" {
		return ErrNo(m.errno)
	}

	return 0
}

func (m *failingReplicationMethods) Undo(conn *SQLiteConn) ErrNo {
	m.fired = append(m.fired, "undo")

	if m.hook == "begin" {
		panic("undo hook should not be reached")
	}
	if m.hook == "undo" {
		return ErrNo(m.errno)
	}

	return 0
}

func (m *failingReplicationMethods) End(conn *SQLiteConn) ErrNo {
	m.fired = append(m.fired, "end")

	if m.hook == "end" {
		return ErrNo(m.errno)
	}

	return 0
}
