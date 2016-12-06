package kvcache

import "time"

type LRUTimeoutMemCache struct {
	maxAge int64
	c      Cache
}

type timeoutWrapper struct {
	expires int64
	value   interface{}
}

func NewLRUTimeoutMemCache(maxBytes, maxAge int) Cache {
	return &LRUTimeoutMemCache{
		maxAge: int64(maxAge),
		c:      NewLRUMemCache(maxBytes),
	}
}

func (c *LRUTimeoutMemCache) Get(
	key string, update UpdateFunc,
) (interface{}, error) {
	now := time.Now().Unix()

	iWrapper, err := c.c.Get(key, func() (interface{}, int, error) {
		value, size, err := update()
		if err != nil {
			return nil, 0, err
		}
		return timeoutWrapper{now + c.maxAge, value}, size, nil
	})

	if err != nil {
		return nil, err
	}

	wrapper := iWrapper.(timeoutWrapper)
	if wrapper.expires < now {
		c.c.Evict(key)
		return c.Get(key, update)
	}

	return wrapper.value, nil
}

func (c LRUTimeoutMemCache) GetStats() (uint64, uint64, uint64) {
	return c.c.GetStats()
}

func (c *LRUTimeoutMemCache) Evict(key string) {
	c.c.Evict(key)
}

func (c *LRUTimeoutMemCache) Clear() {
	c.c.Clear()
}
