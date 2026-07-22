package api

import (
	"fmt"
	"net/http"
	"strings"

	"toggly/backend/internal/auth"
	"toggly/backend/internal/store"
)

type groupResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	System      bool     `json:"system"`
}

func toGroupResponse(g store.Group) groupResponse {
	// g.Permissions is nil for a group with no permissions assigned (the
	// "omitempty" on store.Group.Permissions drops an empty slice entirely
	// when the command is JSON-encoded for the Raft log, so it comes back
	// nil after Apply, even if the original request sent an empty array).
	// Normalize to a non-nil slice here so API clients always see a real
	// array, never null -- mirrors toUserResponse's identical fix for
	// GroupIDs.
	permissions := g.Permissions
	if permissions == nil {
		permissions = []string{}
	}
	return groupResponse{
		ID:          g.ID,
		Name:        g.Name,
		Permissions: permissions,
		System:      g.System,
	}
}

func groupsGetHandler(w http.ResponseWriter, r *http.Request) {
	groups := dataStore.Groups().List()
	resp := make([]groupResponse, 0, len(groups))
	for _, g := range groups {
		resp = append(resp, toGroupResponse(g))
	}
	writeJSON(w, http.StatusOK, map[string]any{"groups": resp})
}

func groupsPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if err := validatePermissions(payload.Permissions); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	group, err := dataStore.Groups().Set(store.Group{
		ID:          store.NewID(),
		Name:        name,
		Permissions: payload.Permissions,
		System:      false, // a client can never create a second system group
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toGroupResponse(group))
}

func groupsPutHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// No pre-check for the Admin group here -- dataStore.Groups().Set below
	// is the single source of truth (System-flag protected) and returns
	// store.ErrProtectedSystemGroup, which writeStoreError maps to the same
	// 403 a duplicated ID check here would have produced.
	existing, ok := dataStore.Groups().Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "group not found"})
		return
	}

	var payload struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if err := validatePermissions(payload.Permissions); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	group, err := dataStore.Groups().Set(store.Group{
		ID:          existing.ID,
		Name:        name,
		Permissions: payload.Permissions,
		System:      false,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toGroupResponse(group))
}

func groupsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// No pre-check for the Admin group here -- see the identical comment in
	// groupsPutHandler; dataStore.Groups().Delete is the single source of
	// truth.
	if _, ok := dataStore.Groups().Get(id); !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "group not found"})
		return
	}

	if err := dataStore.Groups().Delete(id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func validatePermissions(perms []string) error {
	for _, p := range perms {
		if !auth.IsKnownPermission(p) {
			return fmt.Errorf("unknown permission: %q", p)
		}
	}
	return nil
}
