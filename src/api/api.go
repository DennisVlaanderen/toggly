package api

import (
	"encoding/json"
	"net/http"
)

type apiResponse struct {
	Status string `json:"status"`
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/flags", flagsHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, apiResponse{Status: "ok"})
}

func flagsHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"message": "flags endpoint placeholder",
		"flags":   []any{},
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
