package simplelfuda

// LFUDACache is the interface for simple LFUDA cache.
type LFUDACache interface {
	// Adds a value to the cache, returns true if an eviction occurred and
	// updates the "recently used"-ness of the key.
	Set(key, value interface{}) bool

	// Returns key's value from the cache and
	// updates the "recently used"-ness of the key. #value, isFound
	Get(key interface{}) (value interface{}, ok bool)

	// Checks if a key exists in cache without updating the recent-ness.
	Contains(key interface{}) (ok bool)

	// Returns key's value without updating the "recently used"-ness of the key.
	Peek(key interface{}) (value interface{}, ok bool)

	// Removes a key from the cache.
	Remove(key interface{}) bool

	// Returns a slice of the keys in the cache, from oldest to newest.
	Keys() []interface{}

	// Returns the number of items in the cache.
	Len() int

	// Clears all cache entries.
	Purge()
}
