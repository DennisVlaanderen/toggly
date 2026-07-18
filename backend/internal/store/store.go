package store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/raft"
)

// Config describes how to open or bootstrap a Raft-backed Store.
type Config struct {
	NodeID    string
	BindAddr  string
	DataDir   string
	Bootstrap bool
}

// Store is an embedded, Raft-replicated key/value store for feature flags.
type Store struct {
	raft *raft.Raft
	fsm  *fsm
}

// Open starts (or rejoins) a Raft node and returns a ready-to-use Store.
func Open(cfg Config) (*Store, error) {
	r, fsmStore, err := newRaft(cfg.NodeID, cfg.BindAddr, cfg.DataDir, cfg.Bootstrap)
	if err != nil {
		return nil, err
	}
	return &Store{raft: r, fsm: fsmStore}, nil
}

// Get returns the current value of a flag, if it exists.
func (s *Store) Get(key string) (Flag, bool) {
	return s.fsm.get(key)
}

// List returns all known flags.
func (s *Store) List() []Flag {
	return s.fsm.list()
}

// Set applies a flag change through Raft consensus. It only succeeds on the
// cluster leader; with a single bootstrapped node that is always the case.
func (s *Store) Set(flag Flag) (Flag, error) {
	if s.raft.State() != raft.Leader {
		return Flag{}, fmt.Errorf("not the raft leader (leader is %q)", s.raft.Leader())
	}

	data, err := json.Marshal(command{Op: opSet, Flag: flag})
	if err != nil {
		return Flag{}, fmt.Errorf("encode command: %w", err)
	}

	future := s.raft.Apply(data, 5*time.Second)
	if err := future.Error(); err != nil {
		return Flag{}, fmt.Errorf("apply command: %w", err)
	}

	applied, ok := future.Response().(Flag)
	if !ok {
		return Flag{}, fmt.Errorf("unexpected apply response type %T", future.Response())
	}
	return applied, nil
}

// Close shuts down the Raft node.
func (s *Store) Close() error {
	return s.raft.Shutdown().Error()
}
