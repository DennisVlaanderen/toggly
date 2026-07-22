package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"toggly/backend/internal/store"
)

const seedLeaderWaitTimeout = 10 * time.Second

// SeedAdminGroupAndUser ensures the protected Admin group and a
// corresponding admin user (from cfg) exist in the store. It's idempotent:
// on repeat boots it does not recreate the Admin group, and it never
// resets an existing admin's password -- a password changed later through
// the UI must survive restarts.
//
// A freshly bootstrapped single-node cluster isn't leader the instant
// store.Open returns -- the first Raft election takes a heartbeat timeout
// or two -- so this retries on store.ErrNotLeader for a bounded window
// before giving up. On a genuine multi-node follower that never becomes
// leader, it logs and returns nil once the window elapses: the leader will
// have already seeded the state, which then replicates here.
func SeedAdminGroupAndUser(s *store.Store, cfg AdminConfig) error {
	deadline := time.Now().Add(seedLeaderWaitTimeout)
	for {
		err := seedAdminGroupAndUserOnce(s, cfg)
		if err == nil || !errors.Is(err, store.ErrNotLeader) {
			return err
		}
		if time.Now().After(deadline) {
			log.Printf("auth: giving up waiting to become raft leader to seed admin account: %v", err)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func seedAdminGroupAndUserOnce(s *store.Store, cfg AdminConfig) error {
	if _, ok := s.Groups().Get(store.AdminGroupID); !ok {
		if _, err := s.Groups().Set(store.Group{ID: store.AdminGroupID, Name: "Admin", System: true}); err != nil {
			return fmt.Errorf("seed admin group: %w", err)
		}
		log.Println("auth: seeded Admin group")
	}

	if _, ok := s.Users().GetByUsername(cfg.Username); ok {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := store.User{
		ID:           store.NewID(),
		Username:     cfg.Username,
		PasswordHash: hash,
		GroupIDs:     []string{store.AdminGroupID},
		Active:       true,
	}
	if _, err := s.Users().Set(admin); err != nil {
		return fmt.Errorf("seed admin user: %w", err)
	}
	log.Printf("auth: seeded admin user %q", cfg.Username)
	return nil
}
