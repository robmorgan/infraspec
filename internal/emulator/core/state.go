package emulator

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type MemoryStateManager struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryStateManager() *MemoryStateManager {
	return &MemoryStateManager{
		data: make(map[string][]byte),
	}
}

func (m *MemoryStateManager) Get(key string, result interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.data[key]
	if !exists {
		return fmt.Errorf("key %s not found", key)
	}

	return json.Unmarshal(data, result)
}

func (m *MemoryStateManager) Set(key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	m.data[key] = data
	return nil
}

func (m *MemoryStateManager) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[key]; !exists {
		return fmt.Errorf("key %s not found", key)
	}

	delete(m.data, key)
	return nil
}

func (m *MemoryStateManager) List(prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keys []string
	for key := range m.data {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

func (m *MemoryStateManager) Exists(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.data[key]
	return exists
}

// Clear removes all data from the state manager.
func (m *MemoryStateManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string][]byte)
}

// Update atomically reads a value, applies an update function, and writes it back.
// The entire read-modify-write operation is protected by a single lock.
func (m *MemoryStateManager) Update(key string, result interface{}, updateFn func() error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read current value
	data, exists := m.data[key]
	if !exists {
		return fmt.Errorf("key %s not found", key)
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	// Apply update function (modifies result in place)
	if err := updateFn(); err != nil {
		return err
	}

	// Write back updated value
	updatedData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal updated value: %w", err)
	}

	m.data[key] = updatedData
	return nil
}