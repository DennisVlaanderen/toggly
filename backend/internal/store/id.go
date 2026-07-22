package store

import (
	uuid "github.com/hashicorp/go-uuid"
)

// NewID returns a random, unique identifier for a new User or Group record.
// hashicorp/go-uuid is already resolved in this module's dependency graph
// (hashicorp/raft depends on it), so this reuses it rather than hand-rolling
// ID generation.
func NewID() string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		panic("store: failed to generate random id: " + err.Error())
	}
	return id
}
