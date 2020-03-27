package simplelfuda

import (
	"container/list"
	"fmt"
)

/*
Differences between LFUDA and regular LFU cache:
  * The cache dynamically "ages" through a global "age" counter
  * Every cache eviction sets the global "age" counter to the evicted item's hits counter,
  * When setting a new item, its "hits" counter should be set to the cache's "age" value
  * When an existing item is updated, its "hits" counter is incremented by 1 to at least "age" + 1.
*/

// EvictCallback is used to get a callback when a LFUDA entry is evicted
type EvictCallback func(key interface{}, value interface{})

type cachePolicy func(element *item, cacheAge float64) float64

// LFUDA is a non-threadsafe fixed size LFU with Dynamic Aging Cache
type LFUDA struct {
	// size of the entire cache in bytes
	size     float64
	currSize float64
	items    map[interface{}]*item
	freqs    *list.List
	onEvict  EvictCallback
	age      float64
	policy   cachePolicy
}

type item struct {
	key         interface{}
	value       interface{}
	size        float64
	hits        float64
	priorityKey float64
	freqNode    *list.Element
}

type listEntry struct {
	entries     map[*item]byte
	priorityKey float64
}

// NewGDSF constructs an LFUDA of the given size in bytes and uses the GDSF eviction policy
func NewGDSF(size float64, onEvict EvictCallback) *LFUDA {
	return &LFUDA{
		size:     size,
		currSize: 0,
		items:    make(map[interface{}]*item),
		freqs:    list.New(),
		onEvict:  onEvict,
		age:      0,
		policy:   gdsfPolicy,
	}
}

// NewLFUDA constructs an LFUDA of the given size in bytes and uses the LFUDA eviction policy
func NewLFUDA(size float64, onEvict EvictCallback) *LFUDA {
	return &LFUDA{
		size:     size,
		currSize: 0,
		items:    make(map[interface{}]*item),
		freqs:    list.New(),
		onEvict:  onEvict,
		age:      0,
		policy:   lfudaPolicy,
	}
}

// Get looks up a key's value from the cache
func (l *LFUDA) Get(key interface{}) (interface{}, bool) {
	if e, ok := l.items[key]; ok {
		l.increment(e)
		return e.value, true
	}

	return nil, false
}

// Peek looks up a key's value from the cache but will not increment the items hit counter
func (l *LFUDA) Peek(key interface{}) (interface{}, bool) {
	if e, ok := l.items[key]; ok {
		return e.value, true
	}
	return nil, false
}

// Set adds a value to the cache.  Returns true if an eviction occurred.
func (l *LFUDA) Set(key interface{}, value interface{}) bool {
	evicted := false
	if e, ok := l.items[key]; ok {
		// value already exists for key.  overwrite
		e.value = value
		l.increment(e)
	} else {
		// check if we need to evict
		// convert to bytes so we can get the size of the value
		numBytes := float64(len([]byte(fmt.Sprintf("%v", value.(interface{})))))

		// check this value will even fit in the cache.  if not just return
		if l.size < numBytes {
			return false
		}

		// evict until there is room for the new item
		for {
			if l.currSize+numBytes > l.size {
				l.evict()
				evicted = true
			} else {
				break
			}
		}

		// value doesn't exist.  insert
		e := new(item)
		e.size = numBytes
		e.key = key
		e.value = value
		l.items[key] = e
		l.currSize += numBytes
		l.increment(e)
	}
	return evicted
}

// Len returns the number of items in the cache.
func (l *LFUDA) Len() int {
	return len(l.items)
}

// Size returns the number of items in the cache.
func (l *LFUDA) Size() float64 {
	return l.currSize
}

func (l *LFUDA) evict() bool {
	if place := l.freqs.Front(); place != nil {
		for entry := range place.Value.(*listEntry).entries {
			// set age to the value of the evicted object
			// cache age should be less than or equal to the minimum key value in the cache
			l.age = entry.priorityKey

			// since entries is a map this is a random key in the lowest frequency node
			l.Remove(entry.key)
			return true
		}
	}
	return false
}

func (l *LFUDA) increment(e *item) {
	oldNode := e.freqNode
	cursor := e.freqNode
	var nextPlace *list.Element

	if cursor == nil {
		// new entry
		nextPlace = l.freqs.Front()
	} else {
		nextPlace = cursor.Next()
	}

	// must update item's hits before updating priorityKey
	e.hits++
	e.priorityKey = l.policy(e, l.age)

	// move up until hits is < next frequency node's
	for {
		// we've reached the back or the point where the next frequency
		// node is greater than the item's hits count.  Either way, create
		// a new frequency node
		if nextPlace == nil || nextPlace.Value.(*listEntry).priorityKey > e.priorityKey {
			// create a new frequency node
			li := new(listEntry)
			li.priorityKey = e.priorityKey
			li.entries = make(map[*item]byte)
			if cursor != nil {
				nextPlace = l.freqs.InsertAfter(li, cursor)
			} else {
				nextPlace = l.freqs.PushFront(li)
			}
			break
		} else if nextPlace.Value.(*listEntry).priorityKey == e.priorityKey {
			// found the right place
			break
		} else if e.priorityKey > nextPlace.Value.(*listEntry).priorityKey {
			// keep searching
			cursor = nextPlace
			nextPlace = cursor.Next()
		}
	}

	// set the right frequency node in the master list
	e.freqNode = nextPlace
	nextPlace.Value.(*listEntry).entries[e] = 1

	// clenaup
	if oldNode != nil {
		// remove from old position
		l.remEntry(oldNode, e)
	}
}

// Purge will completely clear the LFUDA cache
func (l *LFUDA) Purge() {
	for k, v := range l.items {
		if l.onEvict != nil {
			l.onEvict(k, v.value)
		}
		delete(l.items, k)
	}
	l.age = 0
	l.currSize = 0
	l.freqs.Init()
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (l *LFUDA) Contains(key interface{}) (ok bool) {
	_, ok = l.items[key]
	return ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained
func (l *LFUDA) Remove(key interface{}) bool {
	if item, ok := l.items[key]; ok {
		if l.onEvict != nil {
			l.onEvict(item.key, item.value)
		}
		delete(l.items, key)
		l.remEntry(item.freqNode, item)

		// subtract current size of the cache by the size of the evicted item
		l.currSize -= item.size

		return true
	}
	return false
}

func (l *LFUDA) remEntry(place *list.Element, entry *item) {
	entries := place.Value.(*listEntry).entries
	delete(entries, entry)
	if len(entries) == 0 {
		l.freqs.Remove(place)
	}
}

// Keys returns a slice of the keys in the cache ordered by frequency
func (l *LFUDA) Keys() []interface{} {
	keys := make([]interface{}, len(l.items))
	i := 0
	for node := l.freqs.Back(); node != nil; node = node.Prev() {
		for ent := range node.Value.(*listEntry).entries {
			keys[i] = ent.key
			i++
		}
	}
	return keys
}

// Age returns the cache age factor
func (l *LFUDA) Age() float64 {
	return l.age
}

// Ki = Ci * Fi + L where C is set to 1
func lfudaPolicy(element *item, cacheAge float64) float64 {
	return element.hits + cacheAge
}

// Ki = Fi * Ci / Si + L where C is set to 1
func gdsfPolicy(element *item, cacheAge float64) float64 {
	return (element.hits / element.size) + cacheAge
}
