package db

import "sync"

// HashMap is uses Go map to cache items
type HashMap[T CacheEntry] struct {
	currentSize uintptr
	maxSize     uintptr
	entries     map[string]T
	mu          sync.RWMutex
	policy      CachePolicy
}

func (h *HashMap[T]) Get(key string) (T, bool) {
	h.mu.RLocker()
	defer h.mu.RUnlock()
	v, ok := h.entries[key]
	if ok {
		h.policy.RefreshEntry(key)
	}
	return v, ok
}

func (h *HashMap[T]) Set(key string, entry T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	//1) Check if the entry already exists and refresh
	_, ok := h.entries[key]
	if ok {
		h.policy.RefreshEntry(key)
		return
	}

	//2) Need to evict to add new if full
	h.evictExtraEntries(entry.Size())

	//3) now add, apply policy and update size
	h.entries[key] = entry
	h.policy.Add(key)
	h.currentSize += entry.Size()
}

// evictExtraEntries makes enough room for the incoming entry
func (h *HashMap[T]) evictExtraEntries(entrySize uintptr) {
	for h.currentSize+entrySize >= h.maxSize {
		evictKey := h.policy.ChooseEvict()
		h.Delete(evictKey)
	}
}

func (h *HashMap[T]) Delete(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.entries, key)
}

// Iterate safely loops other the items. Can typically be used during clean up
func (h *HashMap[T]) Iterate() <-chan T {
	ch := make(chan T)
	go func() {
		h.mu.RLock()
		defer h.mu.RUnlock()
		for _, entry := range h.entries {
			ch <- entry
		}
		close(ch)
	}()
	return ch
}
