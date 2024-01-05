package policy

import (
	"container/list"
)

type LRU struct {
	queue  *list.List
	lookup map[string]*list.Element
}

// Refresh refreshes an existing element which was already seen by the policy
func (l *LRU) Refresh(key string) {
	q := l.queue
	if ele, ok := l.lookup[key]; ok {
		q.MoveToFront(ele)
	}
}

func (l *LRU) Evict() (evicted string) {
	ele := l.queue.Back()
	if ele != nil {
		l.queue.Remove(ele)
		delete(l.lookup, ele.Value.(string))
		evicted = ele.Value.(string)
	}
	return evicted
}

// Add a new element to the eviction structure
func (l *LRU) Add(key string) {
	q := l.queue
	if ele, ok := l.lookup[key]; ok {
		q.MoveToFront(ele)
		return
	}

	ele := q.PushFront(key)
	l.lookup[key] = ele
}

// Delete removes an element from the list. Note that this is done efficiently with the Go standard list package.
func (l *LRU) Delete(key string) {
	if ele, ok := l.lookup[key]; ok {
		l.queue.Remove(ele)
		delete(l.lookup, key)
	}
}

func NewLRU() *LRU {
	return &LRU{
		queue:  list.New(),
		lookup: make(map[string]*list.Element),
	}
}
