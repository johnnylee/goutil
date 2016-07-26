package kvcache

type UpdateFunc func() (value interface{}, size int, err error)

type Cache interface {
	Get(string, UpdateFunc) (interface{}, error)
	GetStats() (inserts, hits, misses uint64)
	Evict(string)
}
