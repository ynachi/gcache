package policy

import (
	"container/heap"
)

type Item struct {
	key   string
	freq  int
	index int
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

//goland:noinspection GoMixedReceiverTypes
func (pq PriorityQueue) Len() int { return len(pq) }

//goland:noinspection GoMixedReceiverTypes
func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the lowest.
	return pq[i].freq < pq[j].freq
}

//goland:noinspection GoMixedReceiverTypes
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

//goland:noinspection GoMixedReceiverTypes
func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

//goland:noinspection GoMixedReceiverTypes
func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// Update modifies the priority and value of an Item in the queue.
//
//goland:noinspection GoMixedReceiverTypes
func (pq *PriorityQueue) Update(item *Item, priority int) {
	item.freq = priority
	heap.Fix(pq, item.index)
}

type LFU struct {
	frequency     map[string]int
	priorityQueue PriorityQueue
	lookup        map[string]*Item
}

func NewLFU() *LFU {
	return &LFU{
		frequency:     make(map[string]int),
		priorityQueue: make(PriorityQueue, 0),
		lookup:        make(map[string]*Item),
	}
}

func (l *LFU) Add(key string) {
	l.frequency[key]++
	if item, exists := l.lookup[key]; exists {
		l.priorityQueue.Update(item, l.frequency[key])
	} else {
		item := &Item{
			key:  key,
			freq: l.frequency[key],
		}
		heap.Push(&(l.priorityQueue), item)
		l.lookup[key] = item
	}
}

func (l *LFU) Evict() (evicted string) {
	item := heap.Pop(&(l.priorityQueue)).(*Item)
	if item == nil {
		return ""
	}
	delete(l.lookup, item.key)
	delete(l.frequency, item.key)
	return item.key
}

func (l *LFU) Delete(key string) {
	if item, exists := l.lookup[key]; exists {
		heap.Remove(&(l.priorityQueue), item.index)
		delete(l.lookup, key)
		delete(l.frequency, key)
	}
}

func (l *LFU) Refresh(key string) {
	l.Add(key)
}
