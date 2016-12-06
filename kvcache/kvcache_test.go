package kvcache

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

func testRunner(c Cache, t *testing.T, done chan bool) {
	for i := 0; i < 2000000; i++ {
		x := rand.Int31n(50)
		if rand.Float64() < 0.01 {
			x = rand.Int31n(9999)
		}

		key := fmt.Sprintf("key-%v", x)
		expectedVal := []byte("value for " + key)

		val, err := c.Get(key, func() (interface{}, int, error) {
			return expectedVal, len(expectedVal), nil
		})

		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(val.([]byte), expectedVal) {
			t.Fatal(val)
		}
	}
	done <- true
}

func testCache(name string, c Cache, t *testing.T) {
	N := 8
	done := make(chan bool, N)

	for i := 0; i < N; i++ {
		go testRunner(c, t, done)
	}

	for i := 0; i < N; i++ {
		<-done
	}

	inserts, hits, misses := c.GetStats()

	fmt.Println(name)
	fmt.Printf("  Inserts: %v\n", inserts)
	fmt.Printf("  Excess inserts: %v\n", misses-inserts)
	fmt.Printf("  Total calls: %v\n", hits+misses)
	fmt.Printf("  Hit-rate: %.2f%%\n", 100*float64(hits)/float64(hits+misses))
}

func TestCache(t *testing.T) {
	c := NewLRUMemCache(2048)
	testCache("LRU", c, t)
	cache := c.(*LRUMemCache)
	fmt.Println("  Total Bytes: ", cache.totalBytes)

	c = NewLRUTimeoutMemCache(2048, 1)
	testCache("LRUTimeout", c, t)
}
