package quickcache

import (
	"sync"

	"github.com/cespare/xxhash"
)

const (
	segmentCount  = 256
	segmentOpMask = 255
)

// cache
type Cache struct {
	segments [segmentCount]segment
	locks    [segmentCount]sync.Mutex
}


func hashFunc(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func NewCache(size int) *Cache {
	cache := new(Cache)
	for i := 0; i < segmentCount; i++ {
		cache.segments[i] = NewSegment(size/segmentCount, i)
	}
	return cache
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	hash := hashFunc(key)
	idx := hash & segmentOpMask
	c.locks[idx].Lock()
	v, err := c.segments[idx].Get(key)
	c.locks[idx].Unlock()
	return v, err
}

func (c *Cache) Set(key, value []byte) error {
	hash := hashFunc(key)
	idx := hash & segmentOpMask
	c.locks[idx].Lock()
	err := c.segments[idx].Set(key, value)
	c.locks[idx].Unlock()
	return err
}

func (c *Cache) Del() {

}