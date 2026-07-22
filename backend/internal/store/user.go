package store

import (
	"errors"
	"fmt"
	"slices"
)

// ErrUsernameTaken is returned by UserRepository.Set (and, ultimately,
// fsm.applyUser) when the username on the given record already belongs to
// a different user. errors.Is-comparable so callers can map it to a
// specific response.
var ErrUsernameTaken = errors.New("username is already taken")

// ErrLastAdmin is returned by UserRepository.Set/Delete (and, ultimately,
// fsm.applyUser) when an operation would leave the cluster with zero active
// admins -- deleting, deactivating, or stripping Admin group membership
// from the sole remaining one. Mirrors the protection fsm.applyGroup gives
// the Admin group itself (see group.go): without it, the Admin group being
// undeletable is no guarantee at all, since the last account that can
// actually use it could still be removed. errors.Is-comparable so callers
// can map it to a specific response.
var ErrLastAdmin = errors.New("cannot remove the last remaining admin account")

// User is a persisted account record. Hashing and validating passwords is
// the auth package's job -- store itself performs no business rules beyond
// what fsm.Apply enforces (e.g. username uniqueness), the same split
// already used for Flag.
type User struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	PasswordHash []byte   `json:"password_hash,omitempty"`
	GroupIDs     []string `json:"group_ids,omitempty"`
	Active       bool     `json:"active"`
	Version      uint64   `json:"version"`
}

func (f *fsm) applyUser(index uint64, cmd command) interface{} {
	switch cmd.Op {
	case opSet:
		for id, existing := range f.users {
			if id != cmd.User.ID && existing.Username == cmd.User.Username {
				return ErrUsernameTaken
			}
		}
		if f.isSoleActiveAdminLocked(cmd.User.ID) && !isActiveAdmin(*cmd.User) {
			return ErrLastAdmin
		}
		cmd.User.Version = index
		f.users[cmd.User.ID] = *cmd.User
		return *cmd.User
	case opDelete:
		if f.isSoleActiveAdminLocked(cmd.Key) {
			return ErrLastAdmin
		}
		delete(f.users, cmd.Key)
		return nil
	default:
		return fmt.Errorf("unknown command op %q", cmd.Op)
	}
}

func isActiveAdmin(u User) bool {
	return u.Active && slices.Contains(u.GroupIDs, AdminGroupID)
}

// isSoleActiveAdminLocked reports whether id currently is the only active
// admin -- an active member of the Admin group with no other active member
// to fall back on. Caller must hold f.mu.
func (f *fsm) isSoleActiveAdminLocked(id string) bool {
	target, ok := f.users[id]
	if !ok || !isActiveAdmin(target) {
		return false
	}
	for otherID, other := range f.users {
		if otherID != id && isActiveAdmin(other) {
			return false
		}
	}
	return true
}

// isSoleActiveAdmin is the read-locking counterpart of
// isSoleActiveAdminLocked, for use outside of Apply (e.g. a repository's
// fast pre-check before proposing a command to Raft at all).
func (f *fsm) isSoleActiveAdmin(id string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.isSoleActiveAdminLocked(id)
}

func (f *fsm) getUser(id string) (User, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	u, ok := f.users[id]
	return u, ok
}

// getUserByUsername is a linear scan -- fine at the scale a single-cluster
// user store operates at; add an index if that ever stops being true.
func (f *fsm) getUserByUsername(username string) (User, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, u := range f.users {
		if u.Username == username {
			return u, true
		}
	}
	return User{}, false
}

func (f *fsm) listUsers() []User {
	f.mu.RLock()
	defer f.mu.RUnlock()
	users := make([]User, 0, len(f.users))
	for _, u := range f.users {
		users = append(users, u)
	}
	return users
}

// UserRepository provides user operations against the store. Obtain one via
// Store.Users().
type UserRepository struct {
	store *Store
}

// Get returns the current state of a user, if it exists.
func (r UserRepository) Get(id string) (User, bool) {
	return r.store.fsm.getUser(id)
}

// GetByUsername looks up a user by their (unique) username.
func (r UserRepository) GetByUsername(username string) (User, bool) {
	return r.store.fsm.getUserByUsername(username)
}

// List returns all known users.
func (r UserRepository) List() []User {
	return r.store.fsm.listUsers()
}

// Set applies a user create/update through Raft consensus. A duplicate
// username, and any edit that would deactivate or de-admin the sole
// remaining admin, are rejected before they're even proposed to Raft as a
// fast path; fsm.Apply enforces both rules as the ultimate source of truth.
func (r UserRepository) Set(user User) (User, error) {
	if existing, ok := r.store.fsm.getUserByUsername(user.Username); ok && existing.ID != user.ID {
		return User{}, ErrUsernameTaken
	}
	if r.store.fsm.isSoleActiveAdmin(user.ID) && !isActiveAdmin(user) {
		return User{}, ErrLastAdmin
	}

	resp, err := r.store.apply(command{Op: opSet, Entity: entityUser, User: &user})
	if err != nil {
		return User{}, err
	}
	switch v := resp.(type) {
	case User:
		return v, nil
	case error:
		return User{}, v
	default:
		return User{}, fmt.Errorf("unexpected apply response type %T", resp)
	}
}

// Delete removes a user by ID. The sole remaining admin can never be
// deleted -- see ErrLastAdmin.
func (r UserRepository) Delete(id string) error {
	if r.store.fsm.isSoleActiveAdmin(id) {
		return ErrLastAdmin
	}

	resp, err := r.store.apply(command{Op: opDelete, Entity: entityUser, Key: id})
	if err != nil {
		return err
	}
	if respErr, ok := resp.(error); ok {
		return respErr
	}
	return nil
}
