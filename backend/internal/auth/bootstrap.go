package auth

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
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

	// Usernames are stored lowercase (see auth.Service.Authenticate and
	// api.usersPostHandler/usersPutHandler) so a configured
	// TOGGLY_ADMIN_USERNAME of e.g. "Admin" still matches an existing
	// "admin" account instead of seeding a distinct-cased duplicate.
	username := strings.ToLower(strings.TrimSpace(cfg.Username))

	if _, ok := s.Users().GetByUsername(username); ok {
		return nil
	}

	// This is a fresh admin account, not a rename of an existing one --
	// TOGGLY_ADMIN_USERNAME only takes effect on the very first boot for a
	// given username, mirroring how TOGGLY_ADMIN_PASSWORD is never re-applied
	// to an existing account (see the doc comment above). If another admin
	// already exists under a different username, warn loudly: the operator
	// likely intended to rename the admin account, but this creates a
	// second, independent one instead, leaving the original fully active.
	if hasOtherActiveAdmin(s) {
		log.Printf("auth: TOGGLY_ADMIN_USERNAME is %q but at least one other Admin-group account already exists; seeding a new admin account rather than renaming the existing one -- the original account remains active and must be deactivated/removed manually if that wasn't intended", username)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := store.User{
		ID:           store.NewID(),
		Username:     username,
		PasswordHash: hash,
		GroupIDs:     []string{store.AdminGroupID},
		Active:       true,
	}
	if _, err := s.Users().Set(admin); err != nil {
		return fmt.Errorf("seed admin user: %w", err)
	}
	log.Printf("auth: seeded admin user %q", username)
	return nil
}

// hasOtherActiveAdmin reports whether any active user already belongs to
// the Admin group -- used only to decide whether seeding a not-yet-existing
// configured admin username is a fresh bootstrap or a likely-unintended
// second admin account alongside an existing one.
func hasOtherActiveAdmin(s *store.Store) bool {
	for _, u := range s.Users().List() {
		if u.Active && slices.Contains(u.GroupIDs, store.AdminGroupID) {
			return true
		}
	}
	return false
}
