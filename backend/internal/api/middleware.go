package api

import (
	"net/http"
	"slices"
)

// requireRoles wraps next so it only runs for requests carrying a valid
// bearer token whose role is one of roles. Accepting a list (rather than a
// single role) lets a route allow several roles to share an action -- e.g.
// both admin and user can read flags -- without a different guard per
// combination.
func requireRoles(next http.HandlerFunc, roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			return
		}

		user, err := authService.ParseToken(authorization)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}

		if !slices.Contains(roles, user.Role) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}

		next(w, r)
	}
}
