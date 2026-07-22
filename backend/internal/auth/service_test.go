package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"toggly/backend/internal/store"
)

func newTestService(t *testing.T, cfg AdminConfig) (*Service, *store.Store) {
	t.Helper()
	s := store.NewTestStore(t)
	if err := SeedAdminGroupAndUser(s, cfg); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	return NewService("test-secret", s), s
}

func TestAuthenticateSucceedsForSeededAdmin(t *testing.T) {
	service, _ := newTestService(t, DefaultAdminConfig())

	user, err := service.Authenticate("admin", "admin123")
	if err != nil {
		t.Fatalf("expected admin login to succeed: %v", err)
	}
	if user.Username != "admin" {
		t.Fatalf("expected username %q, got %q", "admin", user.Username)
	}
}

func TestAuthenticateUsesConfiguredAdminCredentials(t *testing.T) {
	service, _ := newTestService(t, AdminConfig{Username: "root", Password: "hunter22"})

	if _, err := service.Authenticate("admin", "admin123"); err == nil {
		t.Fatal("expected default admin credentials to be rejected once a custom admin is configured")
	}

	user, err := service.Authenticate("root", "hunter22")
	if err != nil {
		t.Fatalf("expected configured admin login to succeed: %v", err)
	}
	if user.Username != "root" {
		t.Fatalf("expected username %q, got %q", "root", user.Username)
	}
}

func TestAuthenticateFailsForUnknownUsername(t *testing.T) {
	service, _ := newTestService(t, DefaultAdminConfig())

	if _, err := service.Authenticate("unknown", "wrong"); err == nil {
		t.Fatal("expected unknown username to fail")
	}
}

// A plain non-admin user is deactivated here rather than the seeded admin
// -- see the comment on TestParseTokenRejectsAlreadyIssuedTokenAfterDeactivation.
func TestAuthenticateFailsForDeactivatedUser(t *testing.T) {
	service, s := newTestService(t, DefaultAdminConfig())

	member, err := s.Users().Set(store.User{
		ID:           store.NewID(),
		Username:     "member",
		PasswordHash: mustBcryptHash("member-password"),
		Active:       true,
	})
	if err != nil {
		t.Fatalf("create member user: %v", err)
	}

	member.Active = false
	if _, err := s.Users().Set(member); err != nil {
		t.Fatalf("deactivate member: %v", err)
	}

	if _, err := service.Authenticate("member", "member-password"); err == nil {
		t.Fatal("expected deactivated user to fail authentication")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	service, _ := newTestService(t, DefaultAdminConfig())

	user, err := service.Authenticate("admin", "admin123")
	if err != nil {
		t.Fatalf("expected admin login to succeed: %v", err)
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

func TestParseTokenAcceptsBearerPrefix(t *testing.T) {
	service, _ := newTestService(t, DefaultAdminConfig())

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

// A plain non-admin user is deactivated here rather than the seeded admin
// -- token invalidation on deactivation is a property of every account, and
// the seeded admin is the cluster's sole admin, whose deactivation is
// separately (and deliberately) rejected by store.ErrLastAdmin.
func TestParseTokenRejectsAlreadyIssuedTokenAfterDeactivation(t *testing.T) {
	service, s := newTestService(t, DefaultAdminConfig())

	created, err := s.Users().Set(store.User{
		ID:           store.NewID(),
		Username:     "member",
		PasswordHash: mustBcryptHash("member-password"),
		Active:       true,
	})
	if err != nil {
		t.Fatalf("create member user: %v", err)
	}

	user, err := service.Authenticate("member", "member-password")
	if err != nil {
		t.Fatalf("expected member login to succeed: %v", err)
	}
	token, err := service.GenerateToken(user)
	if err != nil {
		t.Fatalf("expected token generation to succeed: %v", err)
	}

	created.Active = false
	if _, err := s.Users().Set(created); err != nil {
		t.Fatalf("deactivate member: %v", err)
	}

	if _, err := service.ParseToken(token); err == nil {
		t.Fatal("expected an already-issued token to stop working immediately after deactivation")
	}
}

func TestSeedAdminGroupAndUserIsIdempotent(t *testing.T) {
	s := store.NewTestStore(t)
	cfg := DefaultAdminConfig()

	if err := SeedAdminGroupAndUser(s, cfg); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := SeedAdminGroupAndUser(s, cfg); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	if got := len(s.Users().List()); got != 1 {
		t.Fatalf("expected exactly 1 admin user after two seed calls, got %d", got)
	}
	if got := len(s.Groups().List()); got != 1 {
		t.Fatalf("expected exactly 1 group after two seed calls, got %d", got)
	}
}

func TestSeedAdminGroupAndUserDoesNotResetExistingPassword(t *testing.T) {
	s := store.NewTestStore(t)
	cfg := DefaultAdminConfig()

	if err := SeedAdminGroupAndUser(s, cfg); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	admin, ok := s.Users().GetByUsername(cfg.Username)
	if !ok {
		t.Fatal("expected seeded admin user to exist")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte("changed-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash new password: %v", err)
	}
	admin.PasswordHash = newHash
	if _, err := s.Users().Set(admin); err != nil {
		t.Fatalf("update admin password: %v", err)
	}

	if err := SeedAdminGroupAndUser(s, cfg); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	service := NewService("test-secret", s)
	if _, err := service.Authenticate(cfg.Username, "changed-password"); err != nil {
		t.Fatalf("expected changed password to survive re-seeding: %v", err)
	}
	if _, err := service.Authenticate(cfg.Username, cfg.Password); err == nil {
		t.Fatal("expected default admin password to no longer work after it was changed")
	}
}
