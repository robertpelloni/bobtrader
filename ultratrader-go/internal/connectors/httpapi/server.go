package httpapi

import (
	"encoding/json"
	"net/http"
)

type Status struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	AccountCount int    `json:"account_count"`
}

func NewHandler(status Status) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "name": status.Name})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		code := http.StatusOK
		if !status.Ready {
			code = http.StatusServiceUnavailable
		}
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(status)
	})
	return mux
}
