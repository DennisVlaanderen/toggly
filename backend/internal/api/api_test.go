package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aerendil/backend/internal/store"
)

func TestIsProductionEnvironment(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"production", true},
		{"Production", true},
		{"PRODUCTION", true},
		{"", false},
		{"development", false},
	}

	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("AERENDIL_ENV", tc.value)
			if got := isProductionEnvironment(); got != tc.want {
				t.Fatalf("isProductionEnvironment() with AERENDIL_ENV=%q = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestJWTSecretFromEnvironmentUsesConfiguredSecret(t *testing.T) {
	t.Setenv("AERENDIL_JWT_SECRET", "a-real-secret")
	t.Setenv("AERENDIL_ENV", "")

	if got := jwtSecretFromEnvironment(); got != "a-real-secret" {
		t.Fatalf("expected configured secret, got %q", got)
	}
}

func TestJWTSecretFromEnvironmentFallsBackInDevelopment(t *testing.T) {
	t.Setenv("AERENDIL_JWT_SECRET", "")
	t.Setenv("AERENDIL_ENV", "")

	if got := jwtSecretFromEnvironment(); got != devJWTSecret {
		t.Fatalf("expected dev fallback secret, got %q", got)
	}
}

func TestWriteStoreErrorHidesInternalDetails(t *testing.T) {
	rec := httptest.NewRecorder()
	writeStoreError(rec, errors.New("boom: /some/internal/path"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "boom") || strings.Contains(rec.Body.String(), "/some/internal/path") {
		t.Fatalf("expected internal error details to not appear in response, got %s", rec.Body.String())
	}
}

func TestWriteStoreErrorMapsUsernameTakenTo409(t *testing.T) {
	rec := httptest.NewRecorder()
	writeStoreError(rec, store.ErrUsernameTaken)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "username is already taken") {
		t.Fatalf("expected a clear username-taken message, got %s", rec.Body.String())
	}
}

func TestMeHandlerReturnsResolvedPrincipal(t *testing.T) {
	mux := newTestMux(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenFor(t))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"permissions"`) || !strings.Contains(rec.Body.String(), `"isAdmin"`) {
		t.Fatalf("expected resolved principal shape in response, got %s", rec.Body.String())
	}
}

func TestMeHandlerRequiresAuthentication(t *testing.T) {
	mux := newTestMux(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with no token, got %d", rec.Code)
	}
}
