// Package db package defines interfaces to implement a caching database.
// It also custom storage backends such as skip lists.
// The caching server supports multiple storage backends along with different eviction policies.
// So it is important to provide a shared behavior for the storage itself and the policies.
// To implement a caching strategy (i.e.: db + eviction), one will have to define a Database and a CachePolicy structs
// which implement the relevant interfaces respectively.
// They will then need to define a third structure which combines both.
// For example, a Cache server using LRU eviction with HashMap as backend.
// It is also possible to just define one structure which implements all the required interfaces.
package db

import (
	"github.com/ynachi/gcache/db/eviction"
	"github.com/ynachi/gcache/gerror"
	"strings"
)

// EvictionPolicy defines how entries are evicted.
type EvictionPolicy interface {
	// Refresh refreshes an existing elements
	Refresh(key string)

	// Evict evicts an item and return it key.
	// An error is returned if the key is not found.
	// This method should not return an error.
	Evict() string

	// Add a new element not already managed by this eviction
	Add(key string)

	// Delete remove an element from the eviction metadata.
	Delete(key string)
}

// CreateEvictionPolicy is a factory method for eviction policies.
func CreateEvictionPolicy(evictionType string) (EvictionPolicy, error) {
	switch strings.ToLower(evictionType) {
	case "lfu":
		return eviction.NewLFU(), nil
	case "lru":
		return eviction.NewLRU(), nil
	default:
		return nil, gerror.ErrEvictionPolicyNotFound
	}
}
