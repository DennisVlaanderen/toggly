package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultTokenTTL = 24 * time.Hour

type User struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type Service struct {
	secretKey []byte
	tokenTTL  time.Duration
}

func NewService(secret string) *Service {
	return &Service{
		secretKey: []byte(secret),
		tokenTTL:  defaultTokenTTL,
	}
}

func (s *Service) Authenticate(username, password string) (*User, error) {
	if username == "admin" && password == "admin123" {
		return &User{Username: "admin", Role: "admin"}, nil
	}
	if username == "user" && password == "user123" {
		return &User{Username: "user", Role: "user"}, nil
	}
	return nil, errors.New("invalid username or password")
}

func (s *Service) GenerateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.Username,
		"role": user.Role,
		"exp": time.Now().Add(s.tokenTTL).Unix(),
		"iat": time.Now().Unix(),
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
