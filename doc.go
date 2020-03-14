// Package lfuda provides a Least Frequently Used with Dynamic Aging cache
//
//  In addition to basic LFU functionality it behaves according to the following logic:
//  - Every cache miss increases "misses"" counter by 1, but only up to the frequency of the top item's key
//    (the item's "freq" counter)
//  - When the cache is at max size, setting a item will only be successful if the misses "misses" counter
//    is >= the least frequency used item's "freq" counter
//  - When setting a new item, its "freq" counter should be set to the current "misses" value
//  - When an existing item is updated, its "freq" counter is incremented by 1 to at least "misses" + 1.
//
// The cache in this package take locks while operating.  Its therefore thread-safe and can be used with multiple goroutines
//
// For use with a single goroutine (to avoid the locking overhead), the simplelfuda package can be used
package lfuda
