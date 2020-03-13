package lfuda

import (
	"sync"

	"github.com/bparli/go-lfuda/simplelfuda"
)

// Cache is a thread-safe fixed size lfuda cache.
type Cache struct {
	lfuda simplelfuda.LFUDACache
	lock  sync.RWMutex
}

// New creates an lfuda of the given size.
func New(size int) *Cache {
	return NewWithEvict(size, nil)
}

// NewWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewWithEvict(size int, onEvicted func(key interface{}, value interface{})) *Cache {
	lfuda := simplelfuda.NewLFUDA(size, simplelfuda.EvictCallback(onEvicted))
	return &Cache{
		lfuda: lfuda,
	}
}

// Purge is used to completely clear the cache.
func (c *Cache) Purge() {
	c.lock.Lock()
	c.lfuda.Purge()
	c.lock.Unlock()
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (c *Cache) Set(key, value interface{}) (evicted bool) {
	c.lock.Lock()
	evicted = c.lfuda.Set(key, value)
	c.lock.Unlock()
	return evicted
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key interface{}) (value interface{}, ok bool) {
	c.lock.Lock()
	value, ok = c.lfuda.Get(key)
	c.lock.Unlock()
	return value, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *Cache) Contains(key interface{}) bool {
	c.lock.RLock()
	containKey := c.lfuda.Contains(key)
	c.lock.RUnlock()
	return containKey
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *Cache) Peek(key interface{}) (value interface{}, ok bool) {
	c.lock.RLock()
	value, ok = c.lfuda.Peek(key)
	c.lock.RUnlock()
	return value, ok
}

// ContainsOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *Cache) ContainsOrSet(key, value interface{}) (ok, evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.lfuda.Contains(key) {
		return true, false
	}
	evicted = c.lfuda.Set(key, value)
	return false, evicted
}

// PeekOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *Cache) PeekOrSet(key, value interface{}) (previous interface{}, ok, evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	previous, ok = c.lfuda.Peek(key)
	if ok {
		return previous, true, false
	}

	evicted = c.lfuda.Set(key, value)
	return nil, false, evicted
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key interface{}) (present bool) {
	c.lock.Lock()
	present = c.lfuda.Remove(key)
	c.lock.Unlock()
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *Cache) Keys() []interface{} {
	c.lock.RLock()
	keys := c.lfuda.Keys()
	c.lock.RUnlock()
	return keys
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.lock.RLock()
	length := c.lfuda.Len()
	c.lock.RUnlock()
	return length
}
