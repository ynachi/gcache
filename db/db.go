// Package db package defines interfaces to implement a caching database.
// It also custom storage backends such as skip lists.
// The caching server supports multiple storage backends along with different eviction policies.
// So it is important to provide a shared behavior for the storage itself and the policies.
// To implement a caching strategy (i.e.: db + policy), one will have to define a Database and a CachePolicy structs
// which implement the relevant interfaces respectively.
// They will then need to define a third structure which combines both.
// For example, a Cache server using LRU policy with HashMap as backend.
// It is also possible to just define one structure which implements all the required interfaces.
package db

import (
	"github.com/ynachi/gcache/db/policy"
	"github.com/ynachi/gcache/gerror"
	"hash/fnv"
	"strings"
	"sync"
)

// Eviction defines how entries are evicted.
type Eviction interface {
	// Refresh refreshes an existing elements
	Refresh(key string)

	// Evict evicts an item and return it key.
	// An error is returned if the key is not found.
	// This method should not return an error.
	Evict() string

	// Add a new element not already managed by this policy
	Add(key string)

	// Delete remove an element from the policy metadata.
	Delete(key string)
}

// CreateEvictionPolicy is a factory method for eviction policies.
func CreateEvictionPolicy(evictionType string) (Eviction, error) {
	switch strings.ToLower(evictionType) {
	case "lfu":
		return policy.NewLFU(), nil
	case "lru":
		return policy.NewLRU(), nil
	default:
		return nil, gerror.ErrEvictionPolicyNotFound
	}
}

type CMap struct {
	mu      sync.Mutex
	buckets []map[string]*Entry
}

func NewCmap(size int) *CMap {
	buckets := make([]map[string]*Entry, 0, size)
	for i := 0; i < size; i++ {
		buckets = append(buckets, make(map[string]*Entry))
	}
	return &CMap{
		buckets: buckets,
	}
}

func (c *CMap) GetBucket(key string) map[string]*Entry {
	// no need for lock as the vector itself does not change
	return c.buckets[hashFnv(key)%len(c.buckets)]
}

func (c *CMap) setEntry(key string, entry *Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GetBucket(key)[key] = entry
}

func (c *CMap) getEntryForKey(key string) *Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.GetBucket(key)[key]
}

func (c *CMap) keyIn(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.GetBucket(key)[key] != nil
}

func (c *CMap) DeleteEntry(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.GetBucket(key)[key]; ok {
		delete(c.GetBucket(key), key)
		return 1
	}
	return 0
}

func hashFnv(input string) int {
	hasher := fnv.New64a()
	hasher.Write([]byte(input))
	return int(hasher.Sum64())
}
