package store

import (
	"os"
	"testing"
	"time"

	"github.com/hashicorp/raft"
)

// NewTestStore opens a real, single-node Store backed by a temporary
// on-disk data directory and waits for it to become Raft leader before
// returning. It exists so other internal packages (auth, api) can spin up
// a live store for their own tests without duplicating Raft bootstrap
// boilerplate.
func NewTestStore(t testing.TB) *Store {
	t.Helper()

	dataDir, err := os.MkdirTemp("", "aerendil-test-store-*")
	if err != nil {
		t.Fatalf("create temp data dir: %v", err)
	}
	// Store.Close doesn't close the underlying BoltDB file handle, so on
	// Windows a t.TempDir()-style automatic cleanup can fail the test by
	// racing an open file lock. Best-effort removal avoids failing tests
	// over that pre-existing, unrelated cleanup gap.
	t.Cleanup(func() {
		_ = os.RemoveAll(dataDir)
	})

	s, err := Open(Config{
		NodeID:    "test-node",
		BindAddr:  "127.0.0.1:0",
		DataDir:   dataDir,
		Bootstrap: true,
	})
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if s.raft.State() == raft.Leader {
			return s
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for raft node to become leader")
	return nil
}
