package store

import "fmt"

// Flag is a single feature flag record replicated across the cluster.
type Flag struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
	Value   string `json:"value,omitempty"`
	Version uint64 `json:"version"`
}

func (f *fsm) applyFlag(index uint64, cmd command) interface{} {
	switch cmd.Op {
	case opSet:
		cmd.Flag.Version = index
		f.flags[cmd.Flag.Key] = *cmd.Flag
		return *cmd.Flag
	case opDelete:
		delete(f.flags, cmd.Key)
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

// FlagRepository provides flag operations against the store. Obtain one via
// Store.Flags(); it's a stateless one-field wrapper, cheap to construct on
// every call, so it never needs to be cached or stored as a field.
type FlagRepository struct {
	store *Store
}

// Get returns the current value of a flag, if it exists.
func (r FlagRepository) Get(key string) (Flag, bool) {
	return r.store.fsm.get(key)
}

// List returns all known flags.
func (r FlagRepository) List() []Flag {
	return r.store.fsm.list()
}

// Set applies a flag change through Raft consensus. It only succeeds on the
// cluster leader; with a single bootstrapped node that is always the case.
func (r FlagRepository) Set(flag Flag) (Flag, error) {
	resp, err := r.store.apply(command{Op: opSet, Entity: entityFlag, Flag: &flag})
	if err != nil {
		return Flag{}, err
	}
	applied, ok := resp.(Flag)
	if !ok {
		return Flag{}, fmt.Errorf("unexpected apply response type %T", resp)
	}
	return applied, nil
}
