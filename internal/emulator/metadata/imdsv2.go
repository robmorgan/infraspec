package metadata

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

const (
	// MinTokenTTL is the minimum TTL for IMDSv2 tokens (1 second)
	MinTokenTTL = 1
	// MaxTokenTTL is the maximum TTL for IMDSv2 tokens (6 hours)
	MaxTokenTTL = 21600
	// DefaultTokenTTL is the default TTL if not specified (6 hours)
	DefaultTokenTTL = 21600
)

// GenerateToken creates a new IMDSv2 session token
func GenerateToken(state emulator.StateManager, ttl int) (*IMDSv2Token, error) {
	// Validate TTL
	if ttl < MinTokenTTL || ttl > MaxTokenTTL {
		return nil, fmt.Errorf("invalid TTL: must be between %d and %d seconds", MinTokenTTL, MaxTokenTTL)
	}

	// Generate a random 56-character token (similar to AWS)
	tokenValue := generateTokenString()

	// Calculate expiration
	expiresAt := time.Now().Add(time.Duration(ttl) * time.Second)

	token := &IMDSv2Token{
		Token:     tokenValue,
		ExpiresAt: expiresAt,
		TTL:       ttl,
	}

	// Store token in state
	key := fmt.Sprintf("metadata:tokens:%s", tokenValue)
	if err := state.Set(key, token); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return token, nil
}

// ValidateToken checks if a token is valid and not expired
func ValidateToken(state emulator.StateManager, tokenValue string) (bool, error) {
	if tokenValue == "" {
		return false, nil
	}

	key := fmt.Sprintf("metadata:tokens:%s", tokenValue)
	var token IMDSv2Token
	if err := state.Get(key, &token); err != nil {
		// Token not found
		return false, nil
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		// Clean up expired token
		_ = state.Delete(key)
		return false, nil
	}

	return true, nil
}

// ParseTTL parses the TTL from the header value
func ParseTTL(ttlHeader string) (int, error) {
	if ttlHeader == "" {
		return DefaultTokenTTL, nil
	}

	ttl, err := strconv.Atoi(ttlHeader)
	if err != nil {
		return 0, fmt.Errorf("invalid TTL value: %w", err)
	}

	if ttl < MinTokenTTL || ttl > MaxTokenTTL {
		return 0, fmt.Errorf("TTL must be between %d and %d seconds", MinTokenTTL, MaxTokenTTL)
	}

	return ttl, nil
}

// generateTokenString generates a random token string similar to AWS IMDSv2 tokens
func generateTokenString() string {
	// AWS tokens are typically 56 characters
	// We'll generate using UUIDs for simplicity
	part1 := uuid.New().String()
	part2 := uuid.New().String()
	// Remove hyphens and combine
	token := part1[:8] + part2[:8] + uuid.New().String()[:8] + uuid.New().String()[:8] + uuid.New().String()[:8] + uuid.New().String()[:8] + uuid.New().String()[:8]
	return token
}
