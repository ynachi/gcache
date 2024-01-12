package db

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
	channel     chan Ops
	maxItems    int
	currentSize int
	storage     map[string]*Entry
	eviction    EvictionPolicy
}

// get retrieves the value associated with the provided key from the cache.
// It returns the value along with a boolean flag indicating if the value was found in the cache or not.
func (c *Cache) get(key string) (string, bool) {
	e, ok := c.storage[key]
	if ok {
		c.eviction.Refresh(e.key)
		return e.value, ok
	}
	return "", false
}

func (c *Cache) Get(key string) (string, bool) {
	cmd := Ops{
		Name:   "Get",
		Key:    key,
		Result: make(chan opsResult),
	}
	c.channel <- cmd
	result := <-cmd.Result
	return result.stringValue, result.boolValue
}

func (c *Cache) set(key string, value string) {
	e, ok := c.storage[key]
	if ok {
		e.value = value
		c.eviction.Refresh(e.key)
	} else {
		if c.currentSize >= c.maxItems {
			evictKey := c.eviction.Evict()
			if evictKey != "" {
				delete(c.storage, evictKey)
				c.currentSize -= 1
			}
		}
		c.storage[key] = NewEntry(key, value)
		c.eviction.Add(key)
		c.currentSize += 1
	}
}

func (c *Cache) Set(key string, value string) {
	cmd := Ops{
		Name:  "Set",
		Key:   key,
		Value: value,
	}
	c.channel <- cmd
}

// Delete delete keys and return the number of removed keys
func (c *Cache) delete(keys ...string) int {
	deletedKeys := 0
	for _, key := range keys {
		e, ok := c.storage[key]
		if ok {
			delete(c.storage, key)
			c.currentSize -= 1
			c.eviction.Delete(e.key)
			deletedKeys += 1
		}
	}
	return deletedKeys
}

func (c *Cache) Delete(keys ...string) int {
	cmd := Ops{
		Name:   "Delete",
		Keys:   keys,
		Result: make(chan opsResult),
	}
	c.channel <- cmd
	result := <-cmd.Result
	return result.intValue
}

func NewCache(maxItem int, evictionPolicyType string) (*Cache, error) {
	evictionPolicy, err := CreateEvictionPolicy(evictionPolicyType)
	if err != nil {
		return nil, err
	}
	cache := &Cache{
		maxItems:    maxItem,
		storage:     make(map[string]*Entry),
		currentSize: 0,
		eviction:    evictionPolicy,
	}
	go cache.runWorker()
	return cache, nil
}

/*
  COMMAND TYPE
*/

type opsResult struct {
	stringValue string
	intValue    int
	boolValue   bool
}

type Ops struct {
	// Database ops name
	Name   string
	Key    string
	Keys   []string // for commands which operate on multiple keys.
	Value  string
	Result chan opsResult
}

func (c *Cache) runWorker() {
	for cmd := range c.channel {
		switch cmd.Name {
		case "Set":
			c.set(cmd.Key, cmd.Value)
		case "Get":
			res, ok := c.get(cmd.Key)
			cmd.Result <- opsResult{stringValue: res, boolValue: ok}
		case "Delete":
			res := c.delete(cmd.Keys...)
			cmd.Result <- opsResult{intValue: res}
		}
	}
}
