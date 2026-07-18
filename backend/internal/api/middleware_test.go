package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"toggly/backend/internal/auth"
	"toggly/backend/internal/store"
)

// newTestMux boots a real (single-node, temp-dir-backed) flag store and
// wires it up through RegisterRoutes, then waits for the node to become
// raft leader by retrying a harmless Set until it succeeds.
func newTestMux(t *testing.T) *http.ServeMux {
	t.Helper()

	dataDir, err := os.MkdirTemp("", "toggly-middleware-test-*")
	if err != nil {
		t.Fatalf("create temp data dir: %v", err)
	}
	// store.Close doesn't close the underlying BoltDB file handle, so on
	// Windows a t.TempDir()-style cleanup can race an open file lock. Best
	// effort removal here avoids failing this test over that pre-existing,
	// unrelated cleanup gap.
	t.Cleanup(func() {
		_ = os.RemoveAll(dataDir)
	})

	s, err := store.Open(store.Config{
		NodeID:    "middleware-test",
		BindAddr:  "127.0.0.1:0",
		DataDir:   dataDir,
		Bootstrap: true,
	})
	if err != nil {
		t.Fatalf("open flag store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := s.Set(store.Flag{Key: "warmup", Enabled: true}); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for raft node to become leader: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, s)
	return mux
}

func tokenFor(t *testing.T, role, username string) string {
	t.Helper()
	token, err := authService.GenerateToken(&auth.User{Username: username, Role: role})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func TestFlagsGetRequiresAuthentication(t *testing.T) {
	mux := newTestMux(t)

	req := httptest.NewRequest(http.MethodGet, "/api/flags", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with no token, got %d", rec.Code)
	}
}

func TestFlagsGetAllowsAnyAuthenticatedRole(t *testing.T) {
	mux := newTestMux(t)

	for _, role := range []string{auth.RoleAdmin, auth.RoleUser} {
		req := httptest.NewRequest(http.MethodGet, "/api/flags", nil)
		req.Header.Set("Authorization", "Bearer "+tokenFor(t, role, "someone"))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for role %q, got %d", role, rec.Code)
		}
	}
}

func TestFlagsPostRequiresAdminRole(t *testing.T) {
	mux := newTestMux(t)

	body, _ := json.Marshal(map[string]any{"key": "checkout", "enabled": true})

	req := httptest.NewRequest(http.MethodPost, "/api/flags", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tokenFor(t, auth.RoleUser, "someone"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for user role, got %d", rec.Code)
	}
}

func TestFlagsPostSucceedsForAdminRole(t *testing.T) {
	mux := newTestMux(t)

	body, _ := json.Marshal(map[string]any{"key": "checkout", "enabled": true, "value": "on"})

	req := httptest.NewRequest(http.MethodPost, "/api/flags", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tokenFor(t, auth.RoleAdmin, "someone"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin role, got %d: %s", rec.Code, rec.Body.String())
	}

	var flag store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flag); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if flag.Key != "checkout" || !flag.Enabled || flag.Value != "on" {
		t.Fatalf("unexpected flag in response: %+v", flag)
	}
}
