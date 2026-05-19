package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend-assignment/internal/models"
	"backend-assignment/internal/store"
)

type RateLimitHandler struct {
	limiter *store.RateLimiter
}

func NewRateLimitHandler(limiter *store.RateLimiter) *RateLimitHandler {
	return &RateLimitHandler{limiter: limiter}
}

func (h *RateLimitHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	var req models.RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}

	req.UserID = strings.TrimSpace(req.UserID)
	if req.UserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user_id is required and cannot be empty or just whitespace"})
		return
	}

	if req.Payload == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "payload is required"})
		return
	}

	allowed := h.limiter.AllowRequest(req.UserID, time.Now())

	w.Header().Set("Content-Type", "application/json")
	if !allowed {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{"error": "Rate limit exceeded. Maximum 5 requests per minute allowed."})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Request accepted"})
}

func (h *RateLimitHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats := h.limiter.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
