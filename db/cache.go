package db

import (
	"sync"
	"sync/atomic"
)

type Entry struct {
	key   string
	value string
}

func NewEntry(key string, value string) *Entry {
	return &Entry{key: key, value: value}
}

// Cache is the storage of our cache server.
// For now, we apply global lock for simplicity.
// Later on, we could average RWLocks and fine-grained locking strategy.
// This means that there is no need for mutexes at the eviction struct side.
type Cache struct {
	mu          sync.Mutex
	maxItems    int64
	currentSize atomic.Int64
	storage     map[string]*Entry
	eviction    Eviction
}

// Size returns the current size of the cache.
// This size is not computed to reduce costs on system calls due to locking/unlocking mutexes.
// We use an atomic variable instead.
func (c *Cache) Size() int64 {
	return c.currentSize.Load()
}

func (c *Cache) increment() {
	c.currentSize.Add(1)
}

func (c *Cache) decrement() {
	c.currentSize.Add(-1)
}

// Get retrieves the value associated with the provided key from the cache.
// It returns the value along with a boolean flag indicating if the value was found in the cache or not.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.storage[key]
	if ok {
		c.eviction.Refresh(e.key)
		return e.value, ok
	}
	return "", false
}

func (c *Cache) Set(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.storage[key]
	if ok {
		e.value = value
		c.eviction.Refresh(e.key)
	} else {
		if c.Size() > c.maxItems {
			evictKey := c.eviction.Evict()
			if evictKey != "" {
				delete(c.storage, evictKey)
				c.decrement()
			}
		}
		c.storage[key] = NewEntry(key, value)
		c.eviction.Add(key)
		c.increment()
	}
}

// Delete delete keys and return the number of removed keys
func (c *Cache) Delete(keys ...string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	deletedKeys := 0
	for _, key := range keys {
		e, ok := c.storage[key]
		if ok {
			delete(c.storage, key)
			c.decrement()
			c.eviction.Delete(e.key)
			deletedKeys += 1
		}
	}
	return deletedKeys
}

func NewCache(maxItem int64, evictionPolicyType string) (*Cache, error) {
	evictionPolicy, err := CreateEvictionPolicy(evictionPolicyType)
	if err != nil {
		return nil, err
	}
	return &Cache{
		maxItems:    maxItem,
		storage:     make(map[string]*Entry),
		currentSize: atomic.Int64{},
		eviction:    evictionPolicy,
	}, nil
}
