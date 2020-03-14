package lfuda

import (
	"math/rand"
	"testing"
)

func BenchmarkLFUDA(b *testing.B) {
	l := New(8192)

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = rand.Int63() % 16384
		} else {
			trace[i] = rand.Int63() % 32768
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Set(trace[i], trace[i])
	}
	var hit, miss int
	for i := 0; i < b.N; i++ {
		_, ok := l.Get(trace[i])
		if ok {
			hit++
		} else {
			miss++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func BenchmarkLFUDA_Rand(b *testing.B) {
	l := New(8192)

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = rand.Int63() % 32768
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			l.Set(trace[i], trace[i])
		}
		if i%7 == 0 {
			for j := 0; j < 20; j++ {
				_, ok := l.Get(trace[i])
				if ok {
					hit++
				} else {
					miss++
				}
			}
		} else {
			_, ok := l.Get(trace[i])
			if ok {
				hit++
			} else {
				miss++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func TestLFUDA(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}) {
		if k != v {
			t.Fatalf("Evict values not equal (%v!=%v)", k, v)
		}
		evictCounter++
	}
	l := NewWithEvict(128, onEvicted)

	numSet := 0
	for i := 0; i < 256; i++ {
		if l.Set(i, i) {
			numSet++
		}
	}
	if l.Len() != 128 {
		t.Fatalf("bad len: %v", l.Len())
	}

	// rest should not have been set since misses is still 0
	if numSet != 128 {
		t.Fatalf("bad evict count: %v", evictCounter)
	}

	for i, k := range l.Keys() {
		if v, ok := l.Get(k); !ok || v != k || v != i {
			t.Fatalf("bad key: %v, %v, %t, %d", k, v, ok, i)
		}
	}

	// bummp the hits counter of each item in cache
	for i := 0; i < 128; i++ {
		_, ok := l.Get(i)
		if !ok {
			t.Fatalf("should not be evicted")
		}
	}

	// these should all be misses
	for i := 128; i < 256; i++ {
		_, ok := l.Get(i)
		if ok {
			t.Fatalf("should not be in cache")
		}
	}

	if ok := l.Set(256, 256); !ok {
		t.Fatalf("Wasn't able to set key/value in cache (but should have been)")
	}

	// expect 256 to be last key in l.Keys()
	for i, k := range l.Keys() {
		if i == 127 && k != 256 {
			t.Fatalf("out of order key: %v", k)
		}
	}

	if val, _ := l.Get(256); val != 256 {
		t.Fatalf("Wrong value returned for key")
	}

	l.Purge()
	if l.Len() != 0 {
		t.Fatalf("bad len: %v", l.Len())
	}
	if _, ok := l.Get(200); ok {
		t.Fatalf("should contain nothing")
	}
}

// test that Set returns true/false
func TestLFUDASet(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}) {
		evictCounter++
	}

	l := NewWithEvict(1, onEvicted)

	if l.Set(1, 1) == false || evictCounter != 0 {
		t.Errorf("should be able to set")
	}
	if l.Set(2, 2) == true || evictCounter != 0 {
		t.Errorf("should not be able to set (yet)")
	}
	// trigger a miss
	l.Get(2)

	// now try setting again
	if l.Set(2, 2) == false || evictCounter != 1 {
		t.Errorf("should be able to set (yet) and an eviction should have occurred")
	}
}

// test that Contains doesn't update recent-ness
func TestLFUDAContains(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}) {
		evictCounter++
	}

	l := NewWithEvict(1, onEvicted)

	if l.Set(1, 1) == false || evictCounter != 0 {
		t.Errorf("should be able to set")
	}
	if l.Set(2, 2) == true || evictCounter != 0 {
		t.Errorf("should not be able to set (yet)")
	}
	// should not trigger a miss
	l.Contains(2)

	// now try setting again
	if l.Set(2, 2) == true || evictCounter != 0 {
		t.Errorf("should not be able to set (yet)")
	}
}
