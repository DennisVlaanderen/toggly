package store

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID returns a random, sufficiently unique identifier for a new User or
// Group record -- plain random bytes hex-encoded, no external UUID
// dependency needed at this repo's scale.
func NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("store: failed to generate random id: " + err.Error())
	}
	return hex.EncodeToString(b)
}
