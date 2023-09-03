package cache

import (
	"sync"
)

// Cache is here to hold data,
// and present it to others in a controlled manner
// (with locks and read locks)
type Cache[K comparable, V any] struct {
	cache map[K]V

	mutex sync.RWMutex
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		cache: make(map[K]V),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	cached, found := c.cache[key]
	if !found {
		return cached, false
	}

	return cached, true
}

func (c *Cache[K, V]) GetAll() []V {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	all := []V{}
	for _, v := range c.cache {
		all = append(all, v)
	}

	return all
}

func (c *Cache[K, V]) Delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, key)
}

func (c *Cache[K, V]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

func (c *Cache[K, V]) Exists(key K) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, found := c.cache[key]
	return found
}
