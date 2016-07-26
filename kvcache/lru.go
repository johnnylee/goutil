package kvcache

import (
	"container/list"
	"sync"
)

type LRUMemCache struct {
	lock       *sync.Mutex
	maxBytes   int
	totalBytes int
	cache      map[string]*list.Element
	ll         *list.List
	inserts    uint64
	hits       uint64
	misses     uint64
}

type lruItem struct {
	key   string
	size  int
	value interface{}
}

func NewLRUMemCache(maxBytes int) Cache {
	return &LRUMemCache{
		lock:     &sync.Mutex{},
		maxBytes: maxBytes,
		cache:    make(map[string]*list.Element),
		ll:       list.New(),
	}
}

func (c *LRUMemCache) Get(key string, update UpdateFunc) (interface{}, error) {
	c.lock.Lock()
	if val, ok := c.get(key); ok {
		c.hits++
		c.lock.Unlock()
		return val, nil
	}

	c.misses++

	// Not in cache. Call update without lock.
	c.lock.Unlock()

	newVal, newSize, err := update()
	if err != nil {
		return nil, err
	}

	// Lock again for writing.
	c.lock.Lock()
	defer c.lock.Unlock()

	// Could be in cache now if we lost the update race.
	if val, ok := c.get(key); ok {
		return val, nil
	}

	// If the new value is too large, we won't cache it.
	if newSize > (c.maxBytes / 2) {
		return newVal, nil
	}

	newItem := lruItem{key, newSize, newVal}

	// Update total size.
	c.totalBytes += newSize

	// Evict items until size is acceptable.
	for c.totalBytes > c.maxBytes {
		el := c.ll.Back()
		item := el.Value.(*lruItem)
		c.ll.Remove(el)
		delete(c.cache, item.key)
		c.totalBytes -= item.size
	}

	// Insert.
	c.inserts++
	c.cache[key] = c.ll.PushFront(&newItem)

	// Return new value.
	return newVal, nil
}

func (c *LRUMemCache) get(key string) (interface{}, bool) {
	le, ok := c.cache[key]
	if ok {
		c.ll.MoveToFront(le)
		return le.Value.(*lruItem).value, true
	}
	return nil, false
}

func (c LRUMemCache) GetStats() (uint64, uint64, uint64) {
	return c.inserts, c.hits, c.misses
}

func (c *LRUMemCache) Evict(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	el, ok := c.cache[key]
	if !ok {
		return
	}

	item := el.Value.(*lruItem)
	delete(c.cache, key)
	c.ll.Remove(el)
	c.totalBytes -= item.size
}
