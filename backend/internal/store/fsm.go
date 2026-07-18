package store

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

const (
	opSet    = "set"
	opDelete = "delete"
)

type command struct {
	Op   string `json:"op"`
	Flag Flag   `json:"flag"`
}

// fsm is the Raft finite state machine: an in-memory map of flags, made
// durable by Raft's own replicated log plus periodic snapshots. There's no
// need for the FSM's own storage to be a separate embedded database -- Raft
// already gives us a durable, replicated log to reconstruct it from.
type fsm struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

func newFSM() *fsm {
	return &fsm{flags: make(map[string]Flag)}
}

// Apply is called once per committed log entry, in log order, on every node.
func (f *fsm) Apply(log *raft.Log) interface{} {
	var cmd command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("decode command: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Op {
	case opSet:
		cmd.Flag.Version = log.Index
		f.flags[cmd.Flag.Key] = cmd.Flag
		return cmd.Flag
	case opDelete:
		delete(f.flags, cmd.Flag.Key)
		return nil
	default:
		return fmt.Errorf("unknown command op %q", cmd.Op)
	}
}

func (f *fsm) get(key string) (Flag, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	flag, ok := f.flags[key]
	return flag, ok
}

func (f *fsm) list() []Flag {
	f.mu.RLock()
	defer f.mu.RUnlock()
	flags := make([]Flag, 0, len(f.flags))
	for _, flag := range f.flags {
		flags = append(flags, flag)
	}
	return flags
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	flags := make(map[string]Flag, len(f.flags))
	for k, v := range f.flags {
		flags[k] = v
	}
	return &fsmSnapshot{flags: flags}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	flags := make(map[string]Flag)
	if err := json.NewDecoder(rc).Decode(&flags); err != nil && err != io.EOF {
		return fmt.Errorf("decode snapshot: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.flags = flags
	return nil
}

type fsmSnapshot struct {
	flags map[string]Flag
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	if err := json.NewEncoder(sink).Encode(s.flags); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *fsmSnapshot) Release() {}
