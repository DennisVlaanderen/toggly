package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const defaultTokenTTL = 24 * time.Hour

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// AdminConfig describes the admin account to bootstrap the service with.
type AdminConfig struct {
	Username string
	Password string
}

// DefaultAdminConfig returns the insecure admin/admin123 pair used when no
// admin credentials are configured, so local deploys work out of the box.
func DefaultAdminConfig() AdminConfig {
	return AdminConfig{Username: "admin", Password: "admin123"}
}

type Service struct {
	secretKey         []byte
	tokenTTL          time.Duration
	adminUsername     string
	adminPasswordHash []byte
}

func NewService(secret string, admin AdminConfig) *Service {
	hash, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("auth: failed to hash configured admin password: %v", err))
	}

	return &Service{
		secretKey:         []byte(secret),
		tokenTTL:          defaultTokenTTL,
		adminUsername:     admin.Username,
		adminPasswordHash: hash,
	}
}

func (s *Service) Authenticate(username, password string) (*User, error) {
	if username == s.adminUsername {
		if err := bcrypt.CompareHashAndPassword(s.adminPasswordHash, []byte(password)); err != nil {
			return nil, errors.New("invalid username or password")
		}
		return &User{Username: s.adminUsername, Role: RoleAdmin}, nil
	}
	if username == "user" && password == "user123" {
		return &User{Username: "user", Role: RoleUser}, nil
	}
	return nil, errors.New("invalid username or password")
}

func (s *Service) GenerateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.Username,
		"role": user.Role,
		"exp":  time.Now().Add(s.tokenTTL).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

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

	username, _ := claims["sub"].(string)
	role, _ := claims["role"].(string)
	if username == "" {
		return nil, errors.New("missing username claim")
	}

	return &User{Username: username, Role: role}, nil
}
