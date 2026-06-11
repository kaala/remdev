package handler

import (
	"encoding/json"
	"net/http"

	"remdev/config"
)

type ConfigHandler struct {
	cfg *config.Config
}

func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{cfg: cfg}
}

func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h.cfg)
	case http.MethodPut:
		h.put(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ConfigHandler) put(w http.ResponseWriter, r *http.Request) {
	var updates config.Config
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if updates.Port != 0 {
		h.cfg.Port = updates.Port
	}
	if updates.Host != "" {
		h.cfg.Host = updates.Host
	}
	if updates.Token != "" {
		h.cfg.Token = updates.Token
	}
	if updates.Theme != "" {
		h.cfg.Theme = updates.Theme
	}
	if updates.FontFamily != "" {
		h.cfg.FontFamily = updates.FontFamily
	}
	if updates.FontSize != 0 {
		h.cfg.FontSize = updates.FontSize
	}

	if err := h.cfg.Save(); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.cfg)
}
