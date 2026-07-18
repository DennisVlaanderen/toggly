package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"toggly/backend/internal/auth"
	"toggly/backend/internal/store"
)

type apiResponse struct {
	Status string `json:"status"`
}

const devJWTSecret = "toggly-dev-secret"

var authService = auth.NewService(jwtSecretFromEnvironment(), adminConfigFromEnvironment())
var flagStore *store.Store

func jwtSecretFromEnvironment() string {
	secret := strings.TrimSpace(os.Getenv("TOGGLY_JWT_SECRET"))
	if secret == "" {
		log.Println("TOGGLY_JWT_SECRET not set; using insecure development default")
		return devJWTSecret
	}
	return secret
}

func adminConfigFromEnvironment() auth.AdminConfig {
	defaults := auth.DefaultAdminConfig()

	username := strings.TrimSpace(os.Getenv("TOGGLY_ADMIN_USERNAME"))
	if username == "" {
		username = defaults.Username
	}

	password := os.Getenv("TOGGLY_ADMIN_PASSWORD")
	if strings.TrimSpace(password) == "" {
		log.Println("TOGGLY_ADMIN_PASSWORD not set; using insecure development default")
		password = defaults.Password
	}

	return auth.AdminConfig{Username: username, Password: password}
}

func RegisterRoutes(mux *http.ServeMux, flags *store.Store) {
	flagStore = flags

	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("GET /api/flags", requireRoles(flagsGetHandler, auth.RoleAdmin, auth.RoleUser))
	mux.HandleFunc("POST /api/flags", requireRoles(flagsPostHandler, auth.RoleAdmin))
	mux.HandleFunc("/api/auth/login", loginHandler)
	mux.HandleFunc("/api/auth/me", meHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, apiResponse{Status: "ok"})
}

func flagsGetHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"flags": flagStore.List()})
}

func flagsPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Key     string `json:"key"`
		Enabled bool   `json:"enabled"`
		Value   string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(payload.Key) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key is required"})
		return
	}

	flag, err := flagStore.Set(store.Flag{Key: payload.Key, Enabled: payload.Enabled, Value: payload.Value})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := authService.Authenticate(payload.Username, payload.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		return
	}

	token, err := authService.GenerateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

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

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
