package kvcache

// UpdateFunc is called if a cache item doesn't exists.  It should return the
// new value, the size of the value (not including key), and an error.
type UpdateFunc func() (value interface{}, size int, err error)

type Cache interface {
	Get(string, UpdateFunc) (interface{}, error)
	GetStats() (inserts, hits, misses uint64)
	Evict(string)
	Clear()
}
