package db

import (
	"sync"
	"testing"
)

type MockCacheEntry struct {
	size uintptr
}

func (m MockCacheEntry) Size() uintptr {
	return m.size
}

type MockCachePolicy struct {
	keys []string
}

func (m *MockCachePolicy) Add(key string) {
	m.keys = append(m.keys, key)
}

func (m *MockCachePolicy) RefreshEntry(key string) {
	// Implementation intentionally left out
}

func (m *MockCachePolicy) ChooseEvict() string {
	key := m.keys[0]
	m.keys = m.keys[1:]
	return key
}

func TestGet(t *testing.T) {
	mockPolicy := &MockCachePolicy{}
	h := HashMap[MockCacheEntry]{
		currentSize: 0,
		maxSize:     10,
		entries:     make(map[string]MockCacheEntry),
		mu:          sync.RWMutex{},
		policy:      mockPolicy,
	}
	entry := MockCacheEntry{size: 1}
	key := "key1"
	h.Set(key, entry)

	tests := []struct {
		name           string
		inputKey       string
		expectedExists bool
		expectedEntry  MockCacheEntry
	}{
		{"Existing key", "key1", true, entry},
		{"Non-existing key", "key2", false, MockCacheEntry{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, exists := h.Get(tc.inputKey)
			if exists != tc.expectedExists {
				t.Errorf("Expected existence is %v, but got %v", tc.expectedExists, exists)
			}
			if exists && result != tc.expectedEntry {
				t.Errorf("Expected entry is %v, but got %v", tc.expectedEntry, result)
			}
		})
	}
}