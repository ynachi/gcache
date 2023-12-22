package db

// Database is the general interface for different cache backend.
// Any will typically be a pointer to a cache entry: example *CacheEntry
type Database[T CacheEntry] interface {
	Get(Key string) (T, bool)
	Set(key string, entry T)
	Delete(key string)
	Iterate() <-chan T
}

// CachePolicy defines how entries are evicted.
type CachePolicy interface {
	// RefreshEntry refreshes an existing elements
	RefreshEntry(key string)
	// ChooseEvict returns the key of the item to be evicted.
	// It does not evict the item itself.
	// After the key is selected, we can call the delete method of the backend database to remove the entry.
	ChooseEvict() string

	// Add a new element not already managed by this policy
	Add(key string)
}

type CacheEntry interface {
	Size() uintptr // size in bytes
}

// Define caching entries structs here

// ttlEntry is an entry for lfu policy.
type ttlEntry struct {
	key                 string
	value               string
	expirationTimestamp int64
}

// lfuEntry is an entry for lfu policy.
type lfuEntry struct {
	key         string
	value       string
	accessCount int
}
