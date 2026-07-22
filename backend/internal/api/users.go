package api

import (
	"net/http"
	"slices"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"toggly/backend/internal/store"
)

// minPasswordLength is deliberately simple (length only, no complexity
// rules) -- this is an internal admin tool, not a public consumer app.
const minPasswordLength = 8

// userResponse never includes PasswordHash -- password hashes never leave
// the store/auth layers.
type userResponse struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	GroupIDs []string `json:"groupIds"`
	Active   bool     `json:"active"`
}

func toUserResponse(u store.User) userResponse {
	// u.GroupIDs is nil for a user with no group memberships (the "omitempty"
	// on store.User.GroupIDs drops an empty slice entirely when the command
	// is JSON-encoded for the Raft log, so it comes back nil after Apply,
	// even if the original request sent an empty array). Normalize to a
	// non-nil slice here so API clients always see a real array, never null.
	groupIDs := u.GroupIDs
	if groupIDs == nil {
		groupIDs = []string{}
	}
	return userResponse{
		ID:       u.ID,
		Username: u.Username,
		GroupIDs: groupIDs,
		Active:   u.Active,
	}
}

func usersGetHandler(w http.ResponseWriter, r *http.Request) {
	users := dataStore.Users().List()
	resp := make([]userResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toUserResponse(u))
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": resp})
}

func usersPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		GroupIDs []string `json:"groupIds"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	username := strings.TrimSpace(payload.Username)
	if username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username is required"})
		return
	}
	if payload.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password is required"})
		return
	}
	if len(payload.Password) < minPasswordLength {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	if !requireAdminToGrantAdminGroup(w, r, payload.GroupIDs) {
		return
	}

	// Fast pre-check; fsm.applyUser is the authoritative enforcement point
	// for username uniqueness (see store.ErrUsernameTaken).
	if _, exists := dataStore.Users().GetByUsername(username); exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "username is already taken"})
		return
	}
	for _, groupID := range payload.GroupIDs {
		if _, ok := dataStore.Groups().Get(groupID); !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown group id: " + groupID})
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	user, err := dataStore.Users().Set(store.User{
		ID:           store.NewID(),
		Username:     username,
		PasswordHash: hash,
		GroupIDs:     payload.GroupIDs,
		Active:       true,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

// requireAdminToGrantAdminGroup returns false (having already written a 403
// response) if groupIDs includes the Admin group and the acting principal
// isn't already an admin themselves -- otherwise a users:write-only caller
// could create or edit their way into a full admin account without ever
// holding groups:write or admin rights. Applies whether the Admin group is
// newly granted or merely retained on an edit; a users:write-only caller
// should never be able to touch Admin group membership either way.
func requireAdminToGrantAdminGroup(w http.ResponseWriter, r *http.Request, groupIDs []string) bool {
	if !slices.Contains(groupIDs, store.AdminGroupID) {
		return true
	}
	principal, _ := principalFromContext(r)
	if !principal.IsAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "only an Admin can grant Admin group membership"})
		return false
	}
	return true
}

func usersPutHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	existing, ok := dataStore.Users().Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	var payload struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		GroupIDs []string `json:"groupIds"`
		Active   bool     `json:"active"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	username := strings.TrimSpace(payload.Username)
	if username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username is required"})
		return
	}
	if payload.Password != "" && len(payload.Password) < minPasswordLength {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	if !requireAdminToGrantAdminGroup(w, r, payload.GroupIDs) {
		return
	}

	// Fast pre-check; fsm.applyUser is the authoritative enforcement point
	// for username uniqueness (see store.ErrUsernameTaken). Excludes this
	// user's own record so keeping the same username isn't a false conflict.
	if other, exists := dataStore.Users().GetByUsername(username); exists && other.ID != id {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "username is already taken"})
		return
	}
	for _, groupID := range payload.GroupIDs {
		if _, ok := dataStore.Groups().Get(groupID); !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown group id: " + groupID})
			return
		}
	}

	// An empty password field means "leave it unchanged" -- the edit form
	// never round-trips the existing hash, so this is the only way to
	// distinguish "no change" from "clear the password" (which isn't a
	// supported operation; a password is always required to exist).
	passwordHash := existing.PasswordHash
	if payload.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
			return
		}
		passwordHash = hash
	}

	user, err := dataStore.Users().Set(store.User{
		ID:           existing.ID,
		Username:     username,
		PasswordHash: passwordHash,
		GroupIDs:     payload.GroupIDs,
		Active:       payload.Active,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func usersDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := dataStore.Users().Get(id); !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	if err := dataStore.Users().Delete(id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
