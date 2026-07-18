package store

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/hashicorp/raft"
)

// newTestStore builds a single-node Raft cluster entirely in memory -- no
// disk, no real sockets -- using the same in-memory harness hashicorp/raft's
// own test suite relies on.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	fsmStore := newFSM()
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID("test-node")
	config.HeartbeatTimeout = 50 * time.Millisecond
	config.ElectionTimeout = 50 * time.Millisecond
	config.LeaderLeaseTimeout = 50 * time.Millisecond
	config.CommitTimeout = 5 * time.Millisecond

	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()
	snapshotStore := raft.NewInmemSnapshotStore()
	_, transport := raft.NewInmemTransport("")

	r, err := raft.NewRaft(config, fsmStore, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		t.Fatalf("start raft node: %v", err)
	}

	bootstrapConfig := raft.Configuration{
		Servers: []raft.Server{{ID: config.LocalID, Address: transport.LocalAddr()}},
	}
	if err := r.BootstrapCluster(bootstrapConfig).Error(); err != nil {
		t.Fatalf("bootstrap cluster: %v", err)
	}

	s := &Store{raft: r, fsm: fsmStore}
	t.Cleanup(func() {
		_ = s.Close()
	})

	waitForLeader(t, s)
	return s
}

func waitForLeader(t *testing.T, s *Store) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if s.raft.State() == raft.Leader {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for raft node to become leader")
}

func TestStoreSetAndGet(t *testing.T) {
	s := newTestStore(t)

	applied, err := s.Set(Flag{Key: "new-checkout", Enabled: true, Value: "on"})
	if err != nil {
		t.Fatalf("expected set to succeed: %v", err)
	}
	if applied.Version == 0 {
		t.Fatal("expected applied flag to have a non-zero version")
	}

	got, ok := s.Get("new-checkout")
	if !ok {
		t.Fatal("expected flag to be present after set")
	}
	if !got.Enabled || got.Value != "on" {
		t.Fatalf("unexpected flag state: %+v", got)
	}

	flags := s.List()
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
}

func TestStoreSetOverwritesExistingFlag(t *testing.T) {
	s := newTestStore(t)

	if _, err := s.Set(Flag{Key: "checkout", Enabled: true}); err != nil {
		t.Fatalf("expected first set to succeed: %v", err)
	}
	if _, err := s.Set(Flag{Key: "checkout", Enabled: false}); err != nil {
		t.Fatalf("expected second set to succeed: %v", err)
	}

	got, ok := s.Get("checkout")
	if !ok {
		t.Fatal("expected flag to still be present")
	}
	if got.Enabled {
		t.Fatal("expected the later Set to win")
	}
}

type memSnapshotSink struct {
	buf bytes.Buffer
}

func (s *memSnapshotSink) Write(p []byte) (int, error) { return s.buf.Write(p) }
func (s *memSnapshotSink) Close() error                { return nil }
func (s *memSnapshotSink) ID() string                  { return "test-snapshot" }
func (s *memSnapshotSink) Cancel() error                { return nil }

func (s *memSnapshotSink) reader() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(s.buf.Bytes()))
}

func TestFSMSnapshotRestoreRoundTrip(t *testing.T) {
	f := newFSM()
	f.flags["a"] = Flag{Key: "a", Enabled: true, Version: 1}
	f.flags["b"] = Flag{Key: "b", Enabled: false, Version: 2}

	snap, err := f.Snapshot()
	if err != nil {
		t.Fatalf("expected snapshot to succeed: %v", err)
	}

	sink := &memSnapshotSink{}
	if err := snap.Persist(sink); err != nil {
		t.Fatalf("expected persist to succeed: %v", err)
	}

	restored := newFSM()
	if err := restored.Restore(sink.reader()); err != nil {
		t.Fatalf("expected restore to succeed: %v", err)
	}

	if len(restored.flags) != 2 {
		t.Fatalf("expected 2 restored flags, got %d", len(restored.flags))
	}
	if !restored.flags["a"].Enabled || restored.flags["b"].Enabled {
		t.Fatalf("unexpected restored state: %+v", restored.flags)
	}
}
