package auth

import (
	"errors"
	"sort"

	"toggly/backend/internal/store"
)

const (
	PermFlagsRead   = "flags:read"
	PermFlagsWrite  = "flags:write"
	PermUsersRead   = "users:read"
	PermUsersWrite  = "users:write"
	PermGroupsRead  = "groups:read"
	PermGroupsWrite = "groups:write"
)

// AllPermissions is the catalog of every known permission string --
// surfaced to the Groups UI for building a permission picker, and used to
// validate incoming group.Permissions so a client can't smuggle in an
// unknown or typo'd string. Gating a new endpoint elsewhere is "add one
// const here, reference it in one route registration line" -- nothing else
// in the permission model needs to change.
var AllPermissions = []string{
	PermFlagsRead, PermFlagsWrite,
	PermUsersRead, PermUsersWrite,
	PermGroupsRead, PermGroupsWrite,
}

// IsKnownPermission reports whether perm is one of AllPermissions.
func IsKnownPermission(perm string) bool {
	for _, p := range AllPermissions {
		if p == perm {
			return true
		}
	}
	return false
}

// PermissionSet is a resolved, deduplicated set of permission strings.
type PermissionSet map[string]struct{}

// Has reports whether perm is in the set.
func (p PermissionSet) Has(perm string) bool {
	_, ok := p[perm]
	return ok
}

// Keys returns the permission strings in the set, sorted for stable
// JSON/test output.
func (p PermissionSet) Keys() []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Resolve computes the effective, live permission set for a user: the
// Admin group short-circuits to "all access" (isAdmin=true); otherwise
// perms is the union of all permissions across the user's groups. This is
// called fresh on every permission-gated request -- no caching, nothing
// baked into the JWT -- so group membership/permission changes and
// deactivation take effect on the user's very next request.
func (s *Service) Resolve(userID string) (perms PermissionSet, isAdmin bool, err error) {
	u, ok := s.store.Users().Get(userID)
	if !ok || !u.Active {
		return nil, false, errors.New("user not found or inactive")
	}

	perms = PermissionSet{}
	for _, groupID := range u.GroupIDs {
		g, ok := s.store.Groups().Get(groupID)
		if !ok {
			continue
		}
		if g.ID == store.AdminGroupID && g.System {
			return nil, true, nil
		}
		for _, perm := range g.Permissions {
			perms[perm] = struct{}{}
		}
	}
	return perms, false, nil
}
