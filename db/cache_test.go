package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// define a mock eviction
type mockEvictionPolicy struct {
	accessCount map[string]int
}

func (m *mockEvictionPolicy) Refresh(key string) {
	m.accessCount[key]++
}

func (m *mockEvictionPolicy) Evict() string {
	var keyOfLessCount string
	for key, count := range m.accessCount {
		if keyOfLessCount == "" || count < m.accessCount[keyOfLessCount] {
			keyOfLessCount = key
		}
	}
	delete(m.accessCount, keyOfLessCount)
	return keyOfLessCount
}

func (m *mockEvictionPolicy) Add(key string) {
	m.accessCount[key] = 0
}

func (m *mockEvictionPolicy) Delete(key string) {
	delete(m.accessCount, key)
}

func TestCache(t *testing.T) {
	var testCache = Cache{
		channel:     make(chan Ops),
		maxItems:    5,
		currentSize: 0,
		storage:     make(map[string]*Entry),
		eviction:    &mockEvictionPolicy{make(map[string]int)},
	}
	go testCache.runWorker()

	for i := 0; i < 5; i++ {
		testCache.Set(fmt.Sprintf("key%d", i+1), fmt.Sprintf("value%d", i+1))
	}

	// test each key got the right value
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%d", i+1)
		value, ok := testCache.Get(key)
		assert.Equal(t, fmt.Sprintf("value%d", i+1), value)
		assert.Equal(t, ok, true)
	}

	// Test key update
	testCache.Set("key1", "value6")
	value, ok := testCache.Get("key1")
	assert.Equal(t, "value6", value)

	// Test evicting key with less count
	// Key 2 is now the smallest access count and cache is full, so it should be evicted the next key addition.
	for i := 2; i < 5; i++ {
		testCache.Set(fmt.Sprintf("key%d", i+1), fmt.Sprintf("value%d", i+1))
	}
	testCache.Set("key6", "value6")
	_, ok = testCache.Get("key2")
	assert.Equal(t, false, ok)

	// Test key delete
	key3, ok := testCache.Get("key3")
	assert.Equal(t, "value3", key3)
	assert.Equal(t, true, ok)
	count := testCache.Delete("key3")
	assert.Equal(t, 1, count)
	// key3 should be missing
	_, ok = testCache.Get("key3")
	assert.Equal(t, false, ok)

	// Delete missing key does not crash
	testCache.Delete("key3")
}

func BenchmarkCache(b *testing.B) {
	var testCache = Cache{
		channel:     make(chan Ops),
		maxItems:    1000,
		currentSize: 0,
		storage:     make(map[string]*Entry),
		eviction:    &mockEvictionPolicy{make(map[string]int)},
	}
	go testCache.runWorker()

	rand.NewSource(time.Now().UnixNano())
	keyPrefix := "key" // Prefix for key generation

	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				key := keyPrefix + strconv.Itoa(rand.Intn(100000))
				value := keyPrefix + strconv.Itoa(rand.Intn(100000))

				switch rand.Intn(3) {
				case 0:
					testCache.Set(key, value)
				case 1:
					testCache.Get(key)
				case 2:
					testCache.Delete(key)
				}
			}
		},
	)
}
