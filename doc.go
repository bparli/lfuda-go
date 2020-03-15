// Package lfuda provides a Least Frequently Used with Dynamic Aging cache
//
//  In addition to basic LFU functionality it behaves according to the following logic:
//  - The cache dynamically "ages" through a global "misses" counter
//  - Every cache miss increases the global "misses" counter by 1,
//    but only up to the frequency of the top item's key (the item's "hits" counter)
//  - When the cache is full, setting a item will only be successful if the
//    cache's "misses" counter is >= the least frequently used item's "hits" counter
//  - When setting a new item, its "hits" counter should be set to the cache's "misses" value
//  - When an existing item is updated, its "hits" counter is incremented by 1 to at least "misses" + 1.
//
// The cache in this package take locks while operating.  Its therefore thread-safe and can be used with multiple goroutines
//
// For use with a single goroutine (to avoid the locking overhead), the simplelfuda package can be used
package lfuda
