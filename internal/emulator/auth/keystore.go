package auth

import (
	"errors"
	"sync"
)

var (
	// ErrAccessKeyNotFound is returned when an access key is not found in the store
	ErrAccessKeyNotFound = errors.New("access key not found")
)

// KeyStore defines the interface for managing API keys
type KeyStore interface {
	// GetSecretKey retrieves the secret key for a given access key
	GetSecretKey(accessKey string) (string, error)

	// ValidateAccessKey checks if an access key exists in the store
	ValidateAccessKey(accessKey string) bool

	// AddKey adds a new access key and secret key pair
	AddKey(accessKey, secretKey string)

	// RemoveKey removes an access key from the store
	RemoveKey(accessKey string)
}

// InMemoryKeyStore implements KeyStore using an in-memory map
type InMemoryKeyStore struct {
	mu   sync.RWMutex
	keys map[string]string // accessKey -> secretKey
}

// NewInMemoryKeyStore creates a new in-memory key store
func NewInMemoryKeyStore() *InMemoryKeyStore {
	return &InMemoryKeyStore{
		keys: make(map[string]string),
	}
}

// NewInMemoryKeyStoreWithDefaults creates a key store with default test credentials
func NewInMemoryKeyStoreWithDefaults() *InMemoryKeyStore {
	store := NewInMemoryKeyStore()
	store.AddKey("test", "test")
	return store
}

// GetSecretKey retrieves the secret key for a given access key
func (s *InMemoryKeyStore) GetSecretKey(accessKey string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	secretKey, exists := s.keys[accessKey]
	if !exists {
		return "", ErrAccessKeyNotFound
	}

	return secretKey, nil
}

// ValidateAccessKey checks if an access key exists in the store
func (s *InMemoryKeyStore) ValidateAccessKey(accessKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.keys[accessKey]
	return exists
}

// AddKey adds a new access key and secret key pair
func (s *InMemoryKeyStore) AddKey(accessKey, secretKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[accessKey] = secretKey
}

// RemoveKey removes an access key from the store
func (s *InMemoryKeyStore) RemoveKey(accessKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.keys, accessKey)
}

// Count returns the number of keys in the store
func (s *InMemoryKeyStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.keys)
}
