package auth

import (
	"testing"
)

func TestAuthenticateUser(t *testing.T) {
	service := NewService("test-secret", DefaultAdminConfig())

	user, err := service.Authenticate("admin", "admin123")
	if err != nil {
		t.Fatalf("expected admin login to succeed: %v", err)
	}

	if user.Role != RoleAdmin {
		t.Fatalf("expected admin role, got %q", user.Role)
	}
}

func TestAuthenticateUsesConfiguredAdminCredentials(t *testing.T) {
	service := NewService("test-secret", AdminConfig{Username: "root", Password: "hunter22"})

	if _, err := service.Authenticate("admin", "admin123"); err == nil {
		t.Fatal("expected default admin credentials to be rejected once a custom admin is configured")
	}

	user, err := service.Authenticate("root", "hunter22")
	if err != nil {
		t.Fatalf("expected configured admin login to succeed: %v", err)
	}
	if user.Role != RoleAdmin {
		t.Fatalf("expected admin role, got %q", user.Role)
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	service := NewService("test-secret", DefaultAdminConfig())

	user, err := service.Authenticate("user", "user123")
	if err != nil {
		t.Fatalf("expected user login to succeed: %v", err)
	}

	token, err := service.GenerateToken(user)
	if err != nil {
		t.Fatalf("expected token generation to succeed: %v", err)
	}

	parsed, err := service.ParseToken(token)
	if err != nil {
		t.Fatalf("expected token parsing to succeed: %v", err)
	}

	if parsed.Username != user.Username {
		t.Fatalf("expected username %q, got %q", user.Username, parsed.Username)
	}
}

func TestAuthenticateRejectsInvalidCredentials(t *testing.T) {
	service := NewService("test-secret", DefaultAdminConfig())

	if _, err := service.Authenticate("unknown", "wrong"); err == nil {
		t.Fatal("expected invalid credentials to fail")
	}
}

func TestParseTokenAcceptsBearerPrefix(t *testing.T) {
	service := NewService("test-secret", DefaultAdminConfig())

	user, err := service.Authenticate("admin", "admin123")
	if err != nil {
		t.Fatalf("expected admin login to succeed: %v", err)
	}

	token, err := service.GenerateToken(user)
	if err != nil {
		t.Fatalf("expected token generation to succeed: %v", err)
	}

	parsed, err := service.ParseToken("Bearer " + token)
	if err != nil {
		t.Fatalf("expected bearer token parsing to succeed: %v", err)
	}

	if parsed.Username != user.Username {
		t.Fatalf("expected username %q, got %q", user.Username, parsed.Username)
	}
}
