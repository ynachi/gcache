package db

import (
	"container/list"
	"unsafe"
)

// lruEntry
type lruEntry struct {
	key                 string
	value               string
	expirationTimestamp int64
}

func (e *lruEntry) Key() string {
	return e.key
}

func (e *lruEntry) Value() string {
	return e.key
}

func (e *lruEntry) Size() uintptr {
	return unsafe.Sizeof(e.key) + uintptr(len(e.key)) + unsafe.Sizeof(e.value) + uintptr(len(e.value))
}

type LRUPolicy struct {
	elements list.List
}

// RefreshKey refreshes an existing element which was already seen by the policy
func (l *LRUPolicy) RefreshKey(key string) {
	elem := list.Element{Value: key}
	l.elements.MoveToFront(&elem)
}

func (l *LRUPolicy) ChooseEvict() string {
	keyToEvict := (l.elements.Back().Value).(string)
	l.elements.Remove(l.elements.Back())
	return keyToEvict
}

// Add a new element to the eviction structure
func (l *LRUPolicy) Add(key string) {
	l.elements.PushFront(key)
}
