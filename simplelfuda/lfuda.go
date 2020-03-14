package simplelfuda

import (
	"container/list"
)

/*
Differences between LFUDA and regular LFU cache:
  * Every cache miss increases "misses"" counter by 1, but only up to the frequency of the top item's key
    (the item's "hits" counter)
  * When the cache is at max size, setting a item will only be successful if the misses "misses" counter
    is >= the least frequency used item's "hits" counter
  * When setting a new item, its "hits" counter should be set to the current "misses" value
  * When an existing item is updated, its "freq" counter is incremented by 1 to at least "misses" + 1.
*/

// EvictCallback is used to get a callback when a LFUDA entry is evicted
type EvictCallback func(key interface{}, value interface{})

// LFUDA is a non-threadsafe fixed size LFU with Dynamic Aging Cache
type LFUDA struct {
	size    int
	items   map[interface{}]*item
	freqs   *list.List
	onEvict EvictCallback
	misses  int
}

type item struct {
	key     interface{}
	value   interface{}
	hits    int
	element *list.Element
}

// NewLFUDA constructs an LFUDA of the given size
func NewLFUDA(size int, onEvict EvictCallback) *LFUDA {
	return &LFUDA{
		size:    size,
		items:   make(map[interface{}]*item),
		freqs:   list.New(),
		onEvict: onEvict,
		misses:  0,
	}
}

// Get looks up a key's value from the cache
func (l *LFUDA) Get(key interface{}) (interface{}, bool) {
	if e, ok := l.items[key]; ok {
		l.increment(e)
		return e.value, true
	}
	// only increase misses up to the most hits in the cache
	if len(l.items) > 0 && l.misses < l.freqs.Back().Value.(*item).hits {
		l.misses++
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

// Set adds a value to the cache.  Returns true if value was set, false otherwise.
func (l *LFUDA) Set(key interface{}, value interface{}) bool {
	if e, ok := l.items[key]; ok {
		// value already exists for key.  overwrite
		e.value = value
		l.increment(e)
	} else {
		// check if we need to evict
		if len(l.items) >= l.size {
			// don't evict yet, not until misses > the lowest freq
			if l.misses < l.freqs.Front().Value.(*item).hits {
				return false
			}
			l.evict(1)
		}

		// value doesn't exist.  insert
		e := new(item)
		e.key = key
		e.value = value
		l.items[key] = e
		l.increment(e)
	}
	return true
}

// Len returns the number of items in the cache.
func (l *LFUDA) Len() int {
	return len(l.items)
}

func (l *LFUDA) evict(count int) int {
	var evicted int
	for i := 0; i < count; i++ {
		if elem := l.freqs.Front(); elem != nil {
			entry := elem.Value.(*item)
			delete(l.items, entry.key)
			l.freqs.Remove(elem)
			if l.onEvict != nil {
				l.onEvict(entry.key, entry.value)
			}
		}
		evicted++
	}
	return evicted
}

func (l *LFUDA) increment(e *item) {
	var nextPlace *list.Element
	if e.element == nil {
		// new entry
		e.hits = l.misses + 1
		e.element = l.freqs.PushFront(e)
	} else {
		if e.hits < l.misses {
			e.hits = l.misses
		}
		e.hits++
		for {
			// move up until hits is < next element's
			nextPlace = e.element.Next()
			// we've reached the back
			if nextPlace == nil {
				l.freqs.MoveToBack(e.element)
				break
			} else if e.hits <= nextPlace.Value.(*item).hits {
				break
			} else if e.hits > nextPlace.Value.(*item).hits {
				l.freqs.MoveAfter(e.element, nextPlace.Value.(*item).element)
			}
		}
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
	l.misses = 0
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
		l.freqs.Remove(item.element)
		delete(l.items, key)
		return true
	}
	return false
}

// Keys returns a slice of the keys in the cache ordered by frequency
func (l *LFUDA) Keys() []interface{} {
	keys := make([]interface{}, len(l.items))
	i := 0
	for ent := l.freqs.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*item).key
		i++
	}
	return keys
}
