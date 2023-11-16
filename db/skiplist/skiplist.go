package skiplist

import "math/rand"

type SkipList struct {
	header   *Node
	maxLevel int
}

type Node struct {
	key   int
	nexts []*Node
}

func NewSkipList(maxLevel int) *SkipList {
	header := Node{
		nexts: make([]*Node, maxLevel),
	}
	return &SkipList{
		&header,
		1,
	}
}

// Insert inserts a new key into the skip list
func (s *SkipList) Insert(key int) {
	update := make([]*Node, s.maxLevel+1)
	current := s.header

	// Traverse the list and update the update array
	for i := s.maxLevel; i >= 0; i-- {
		for current.nexts[i] != nil && current.nexts[i].key < key {
			current = current.nexts[i]
		}
		update[i] = current
	}

	// Generate a random level for the new node
	newLevel := randomLevel(s.maxLevel)

	// If the new level is greater than the current level, update the update array
	if newLevel > s.maxLevel {
		for i := s.maxLevel + 1; i <= newLevel; i++ {
			update[i] = s.header
		}
		s.maxLevel = newLevel
	}
	// Create the new node
	newNode := Node{key, newLevel}

	// Update the forward pointers of the new node and its predecessors
	for i := 0; i <= newLevel; i++ {
		newNode.nexts[i] = update[i].nexts[i]
		update[i].nexts[i] = &newNode
	}
}

// randomLevel generates a random level for a new node
func randomLevel(maxLevel int) int {
	level := 0
	for rand.Float64() < 0.5 && level < maxLevel {
		level++
	}
	return level
}

// we can certainly avoid checking for the key equality at each step
// apply this optimization later
func (s *SkipList) Search(key int) *Node {
	curr := s.header
	for i := s.maxLevel - 1; i >= 0; i-- {
		for curr.nexts[i] != nil && curr.nexts[i].key <= key {
			if curr.nexts[i].key == key {
				return curr.nexts[i]
			}
			curr = curr.nexts[i]
		}
	}
	return nil
}

func (s *SkipList) Delete(key int) {}
