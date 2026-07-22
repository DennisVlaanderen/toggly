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

// entity discriminates which map a command applies to.
type entity string

const (
	entityFlag  entity = "flag"
	entityUser  entity = "user"
	entityGroup entity = "group"
)

// command is the single Raft log entry shape for every entity kind this
// store replicates; only the field matching Entity is populated. This is a
// breaking change to the previously flag-only, unversioned command/snapshot
// format -- acceptable pre-v1 with no real deployments to migrate, but it
// does mean an existing on-disk data dir must be wiped before running a
// build that includes this change.
type command struct {
	Op     string `json:"op"`
	Entity entity `json:"entity"`
	Key    string `json:"key,omitempty"`
	Flag   *Flag  `json:"flag,omitempty"`
	User   *User  `json:"user,omitempty"`
	Group  *Group `json:"group,omitempty"`
}

// fsm is the Raft finite state machine: in-memory maps of flags, users, and
// groups, made durable by Raft's own replicated log plus periodic
// snapshots. There's no need for a separate embedded database -- Raft
// already gives us a durable, replicated log to reconstruct all three from.
//
// Each entity's apply logic and read accessors live alongside that
// entity's struct definition (flag.go/user.go/group.go), not here -- this
// file only holds the machinery every entity shares: the command envelope,
// the Apply dispatch switch, and snapshot/restore. Adding a new entity
// means adding a file, not growing this one.
type fsm struct {
	mu     sync.RWMutex
	flags  map[string]Flag
	users  map[string]User
	groups map[string]Group
}

func newFSM() *fsm {
	return &fsm{
		flags:  make(map[string]Flag),
		users:  make(map[string]User),
		groups: make(map[string]Group),
	}
}

// Apply is called once per committed log entry, in log order, on every
// node. Invariants that must hold cluster-wide (e.g. the Admin group's
// immutability) are enforced here rather than only in Store's pre-checks,
// since Apply is the one place guaranteed to run exactly once per entry.
func (f *fsm) Apply(log *raft.Log) interface{} {
	var cmd command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("decode command: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Entity {
	case entityFlag:
		return f.applyFlag(log.Index, cmd)
	case entityUser:
		return f.applyUser(log.Index, cmd)
	case entityGroup:
		return f.applyGroup(log.Index, cmd)
	default:
		return fmt.Errorf("unknown command entity %q", cmd.Entity)
	}
}

// snapshotDoc is the single composite blob a whole FSM snapshot is encoded
// as -- one JSON document covering every entity, not per-entity files,
// following the same "snapshot the whole map" approach the original
// flag-only implementation used.
type snapshotDoc struct {
	Flags  map[string]Flag  `json:"flags"`
	Users  map[string]User  `json:"users"`
	Groups map[string]Group `json:"groups"`
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	doc := snapshotDoc{
		Flags:  make(map[string]Flag, len(f.flags)),
		Users:  make(map[string]User, len(f.users)),
		Groups: make(map[string]Group, len(f.groups)),
	}
	for k, v := range f.flags {
		doc.Flags[k] = v
	}
	for k, v := range f.users {
		doc.Users[k] = v
	}
	for k, v := range f.groups {
		doc.Groups[k] = v
	}
	return &fsmSnapshot{doc: doc}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var doc snapshotDoc
	if err := json.NewDecoder(rc).Decode(&doc); err != nil && err != io.EOF {
		return fmt.Errorf("decode snapshot: %w", err)
	}
	if doc.Flags == nil {
		doc.Flags = make(map[string]Flag)
	}
	if doc.Users == nil {
		doc.Users = make(map[string]User)
	}
	if doc.Groups == nil {
		doc.Groups = make(map[string]Group)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.flags = doc.Flags
	f.users = doc.Users
	f.groups = doc.Groups
	return nil
}

type fsmSnapshot struct {
	doc snapshotDoc
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	if err := json.NewEncoder(sink).Encode(s.doc); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *fsmSnapshot) Release() {}
