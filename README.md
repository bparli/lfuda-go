# lfuda-go
Package lfuda provides a LFU with Dynamic Aging cache library

## Why LFUDA?
There are many LRU and LFU cache policy implementations in go, but not any LFU with Dynamic Aging (as far as I could find).  LFUDA builds upon simple LFU by accommodating shifts in the set of popular objects in the cache.  This only becomes important when the cache is full since previously popular objects could remain for a long time or even indefinitely.  This could prevent newly popular objects from being cached or replacing them.  LFU with Aging policies attempt to address this with a tunable cache age factor to prevent previously popular documents from polluting the cache.

With Dynamic Aging/LFUDA, this is done in a parameter-less way, making it easier to manage compared to LFU with Aging.

## How it works
In addition to basic LFU functionality it behaves according to the following logic:
  * The cache dynamically "ages" through a global "age" counter
  * Depending on the policy used, LFUDA or GreedyDual-Size with Frequency (GDSF), the priority key is determined by the item's frequency and the cache's age.  If using GDSF the item's size is also a factor.
  * Every cache eviction sets the global "age" counter to the evicted item's priority key,
  * When setting a new item, its priority key counter should be set to the cache's "age" value
  * When an existing item is updated, its hits counter is incremented by 1, and its priority key updated depending on the policy.

## Usage
The default cache uses a LFUDA policy, like so:

```go
onEvicted := func(k interface{}, v interface{}) {
  if k != v {
    fmt.Printf("Evicted values (%v: %v)\n", k, v)
  }
}

l := lfuda.NewWithEvict(128, onEvicted)

for i := 0; i < 256; i++ {
  if !l.Set(i, i) {
    fmt.Printf("Setting key/value: %v: %v resulted in eviction\n", i, i)
  }
}

for i := 0; i < 256; i++ {
  if val, ok := l.Get(i); ok {
    fmt.Printf("Key's %v value is %v\n", i, val)
  } else {
    fmt.Printf("Key %v not found in cache\n", i)
  }
}

```

The GDSF policy is also available as an alternative

```go
l := lfuda.NewGDSF(128)

for i := 0; i < 256; i++ {
  if !l.Set(i, i) {
    fmt.Printf("Setting key/value: %v: %v resulted in eviction\n", i, i)
  }
}
```

## Acknowledgements
* Paper outlining LFU with Dynamic Aging [https://www.hpl.hp.com/techreports/98/HPL-98-173.pdf](https://www.hpl.hp.com/techreports/98/HPL-98-173.pdf)
* Squid proxy implementation [https://www.hpl.hp.com/techreports/1999/HPL-1999-69.html](https://www.hpl.hp.com/techreports/1999/HPL-1999-69.html)
* O(1) LFU algorithm paper [http://dhruvbird.com/lfu.pdf](http://dhruvbird.com/lfu.pdf)
* Nice LFU implementation in Go [https://github.com/dgrijalva/lfu-go](https://github.com/dgrijalva/lfu-go)
* Interface patterned after [golang-lru](https://github.com/hashicorp/golang-lru)
