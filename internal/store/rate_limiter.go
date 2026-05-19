package store

import (
	"sync"
	"time"

	"backend-assignment/internal/models"
)

type userState struct {
	acceptedCount    int
	cumulativeReject int
	windowMinute     time.Time
}

type RateLimiter struct {
	mu    sync.RWMutex
	users map[string]*userState
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		users: make(map[string]*userState),
	}
}

// AllowRequest checks if a request is allowed (max 5 per minute per user).
// We use a fixed 1-minute window algorithm by truncating the current time to the minute.
func (r *RateLimiter) AllowRequest(userID string, now time.Time) bool {
	window := now.Truncate(time.Minute)

	r.mu.Lock()
	defer r.mu.Unlock()

	state, exists := r.users[userID]
	if !exists {
		state = &userState{
			acceptedCount:    1,
			cumulativeReject: 0,
			windowMinute:     window,
		}
		r.users[userID] = state
		return true
	}

	// If the recorded window is older than the current window, reset accepted count
	if state.windowMinute.Before(window) {
		state.windowMinute = window
		state.acceptedCount = 0
	}

	if state.acceptedCount >= 5 {
		state.cumulativeReject++
		return false
	}

	state.acceptedCount++
	return true
}

// GetStats returns the stats for all users
// Rejected requests are cumulative, accepted requests are for the current fixed window.
func (r *RateLimiter) GetStats() map[string]models.UserStats {
	now := time.Now().Truncate(time.Minute)
	
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]models.UserStats, len(r.users))
	for userID, state := range r.users {
		// Only show accepted requests for the current window in stats if window matches
		accepted := state.acceptedCount
		if state.windowMinute.Before(now) {
			accepted = 0
		}
		
		stats[userID] = models.UserStats{
			AcceptedRequests: accepted,
			RejectedRequests: state.cumulativeReject,
		}
	}
	return stats
}
