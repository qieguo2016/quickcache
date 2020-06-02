package quickcache

import (
	"fmt"
	"sync"
)

// cache
type Cache struct {
	segments [segmentCount]segment
	locks    [segmentCount]sync.Mutex
}

func NewCache(mbSize int) (*Cache, error) {
	if !isPowerOfTwo(mbSize) {
		return nil, fmt.Errorf("size must be power of two")
	}
	cache := new(Cache)
	size := convertMBToBytes(mbSize)
	for i := 0; i < segmentCount; i++ {
		cache.segments[i] = newSegment(size/segmentCount, i)
	}
	return cache, nil
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	hash := hashFunc(key)
	idx := hash & segmentOpMask
	c.locks[idx].Lock()
	v, err := c.segments[idx].Get(key, hash)
	c.locks[idx].Unlock()
	return v, err
}

func (c *Cache) Set(key, value []byte) error {
	hash := hashFunc(key)
	idx := hash & segmentOpMask
	c.locks[idx].Lock()
	err := c.segments[idx].Set(key, value, hash)
	c.locks[idx].Unlock()
	return err
}

func (c *Cache) Del(key []byte) error {
	hash := hashFunc(key)
	idx := hash & segmentOpMask
	c.locks[idx].Lock()
	err := c.segments[idx].Del(key, hash)
	c.locks[idx].Unlock()
	return err
}
