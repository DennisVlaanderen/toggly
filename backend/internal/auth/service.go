package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"toggly/backend/internal/store"
)

const defaultTokenTTL = 24 * time.Hour

// dummyPasswordHash is compared against on every failed username lookup in
// Authenticate, so an unknown/inactive username still costs exactly one
// bcrypt compare -- without this, response latency alone would reveal
// whether a username exists even though the returned error text is
// identical either way.
var dummyPasswordHash = mustBcryptHash("toggly-timing-safe-dummy-password")

func mustBcryptHash(password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("auth: failed to precompute dummy bcrypt hash: " + err.Error())
	}
	return hash
}

// User is the request-scoped resolved principal returned by Authenticate
// and ParseToken. It's deliberately a different type from store.User,
// which additionally carries PasswordHash/GroupIDs/Active and never leaves
// the store/auth layers.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// AdminConfig describes the admin account seeded into the store at
// bootstrap (see SeedAdminGroupAndUser) -- it is no longer held directly by
// Service, since the admin is just a regular persisted User once seeded.
type AdminConfig struct {
	Username string
	Password string
}

// DefaultAdminConfig returns the insecure admin/admin123 pair used when no
// admin credentials are configured, so local deploys work out of the box.
func DefaultAdminConfig() AdminConfig {
	return AdminConfig{Username: "admin", Password: "admin123"}
}

// Service issues/parses JWTs and authenticates against the persisted user
// store. Permission resolution (Resolve, in permissions.go) also reads
// from the same store, fresh per call.
type Service struct {
	secretKey []byte
	tokenTTL  time.Duration
	store     *store.Store
}

func NewService(secret string, s *store.Store) *Service {
	return &Service{
		secretKey: []byte(secret),
		tokenTTL:  defaultTokenTTL,
		store:     s,
	}
}

func (s *Service) Authenticate(username, password string) (*User, error) {
	u, ok := s.store.Users().GetByUsername(username)
	valid := ok && u.Active

	hash := dummyPasswordHash
	if valid {
		hash = u.PasswordHash
	}
	compareErr := bcrypt.CompareHashAndPassword(hash, []byte(password))

	if !valid || compareErr != nil {
		return nil, errors.New("invalid username or password")
	}
	return &User{ID: u.ID, Username: u.Username}, nil
}

func (s *Service) GenerateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(s.tokenTTL).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ParseToken validates the token's signature/expiry and resolves the
// carried subject against the live user store -- username is never trusted
// from the token itself, so a renamed or deactivated user is reflected
// immediately rather than only after the token expires.
func (s *Service) ParseToken(tokenString string) (*User, error) {
	trimmed := strings.TrimSpace(tokenString)
	if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
		trimmed = strings.TrimSpace(trimmed[7:])
	}

	token, err := jwt.Parse(trimmed, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, _ := claims["sub"].(string)
	if userID == "" {
		return nil, errors.New("missing subject claim")
	}

	u, ok := s.store.Users().Get(userID)
	if !ok || !u.Active {
		return nil, errors.New("user not found or inactive")
	}

	return &User{ID: u.ID, Username: u.Username}, nil
}
