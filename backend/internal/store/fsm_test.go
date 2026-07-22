package store

import (
	"errors"
	"testing"
)

func TestFSMSnapshotRestoreRoundTripAllEntities(t *testing.T) {
	f := newFSM()
	f.flags["a"] = Flag{Key: "a", Enabled: true, Version: 1}
	f.users["u1"] = User{ID: "u1", Username: "alice", GroupIDs: []string{"editors"}, Active: true, Version: 2}
	f.groups["editors"] = Group{ID: "editors", Name: "Editors", Permissions: []string{"flags:read"}, Version: 3}
	f.groups[AdminGroupID] = Group{ID: AdminGroupID, Name: "Admin", System: true, Version: 4}

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

	if len(restored.flags) != 1 || !restored.flags["a"].Enabled {
		t.Fatalf("unexpected restored flags: %+v", restored.flags)
	}

	restoredUser, ok := restored.users["u1"]
	if !ok || restoredUser.Username != "alice" || len(restoredUser.GroupIDs) != 1 {
		t.Fatalf("unexpected restored users: %+v (ok=%v)", restoredUser, ok)
	}

	if len(restored.groups) != 2 {
		t.Fatalf("expected 2 restored groups, got %d", len(restored.groups))
	}
	restoredAdmin, ok := restored.groups[AdminGroupID]
	if !ok || !restoredAdmin.System {
		t.Fatalf("expected Admin group to round-trip with System=true, got %+v (ok=%v)", restoredAdmin, ok)
	}
}

func TestFSMApplyRejectsAdminGroupMutationEvenIfSuccessfullyRaftCommitted(t *testing.T) {
	f := newFSM()
	f.groups[AdminGroupID] = Group{ID: AdminGroupID, Name: "Admin", System: true, Version: 1}

	resp := f.applyGroup(2, command{Op: opSet, Entity: entityGroup, Group: &Group{ID: AdminGroupID, Name: "Renamed"}})
	if _, ok := resp.(error); !ok {
		t.Fatalf("expected applyGroup to return an error for a System group edit, got %+v", resp)
	}

	resp = f.applyGroup(3, command{Op: opDelete, Entity: entityGroup, Key: AdminGroupID})
	if _, ok := resp.(error); !ok {
		t.Fatalf("expected applyGroup to return an error for a System group delete, got %+v", resp)
	}

	if got := f.groups[AdminGroupID]; got.Name != "Admin" {
		t.Fatalf("expected Admin group to be unchanged, got %+v", got)
	}
}

func TestApplyUserRejectsDuplicateUsername(t *testing.T) {
	f := newFSM()
	f.users["u1"] = User{ID: "u1", Username: "alice", Active: true, Version: 1}

	resp := f.applyUser(2, command{Op: opSet, Entity: entityUser, User: &User{ID: "u2", Username: "alice"}})
	respErr, ok := resp.(error)
	if !ok {
		t.Fatalf("expected applyUser to return an error for a duplicate username, got %+v", resp)
	}
	if !errors.Is(respErr, ErrUsernameTaken) {
		t.Fatalf("expected error to be ErrUsernameTaken, got %v", respErr)
	}
	if _, exists := f.users["u2"]; exists {
		t.Fatal("expected the conflicting user to not be written")
	}
}

func TestApplyUserAllowsUpdatingSameUserWithSameUsername(t *testing.T) {
	f := newFSM()
	f.users["u1"] = User{ID: "u1", Username: "alice", Active: true, Version: 1}

	resp := f.applyUser(2, command{Op: opSet, Entity: entityUser, User: &User{ID: "u1", Username: "alice", Active: false}})
	updated, ok := resp.(User)
	if !ok {
		t.Fatalf("expected applyUser to succeed updating the same user, got %+v", resp)
	}
	if updated.Active {
		t.Fatalf("expected the update to be applied, got %+v", updated)
	}
}
