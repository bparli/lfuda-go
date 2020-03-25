package simplelfuda

import (
	"fmt"
	"testing"
)

func TestLFU(t *testing.T) {
	c := NewLFUDA(2, nil)
	c.Set("a", "a")
	if v, _ := c.Get("a"); v != "a" {
		t.Errorf("Value was not saved: %v != 'a'", v)
	}
	if l := c.Len(); l != 1 {
		t.Errorf("Length was not updated: %v != 1", l)
	}

	c.Set("b", "b")
	if v, _ := c.Get("b"); v != "b" {
		t.Errorf("Value was not saved: %v != 'b'", v)
	}
	if l := c.Len(); l != 2 {
		t.Errorf("Length was not updated: %v != 2", l)
	}

	if v, ok := c.Get("b"); !ok {
		t.Errorf("Value was improperly evicted: %v != 'b'", v)
	}

	if ok := c.Remove("a"); !ok {
		t.Errorf("Item was not removed: a")
	}
	if v, _ := c.Get("a"); v != nil {
		t.Errorf("Value was not removed: %v", v)
	}
	if l := c.Len(); l != 1 {
		t.Errorf("Length was not updated: %v != 1", l)
	}
}

func TestCacheSize(t *testing.T) {
	// 10 bytes total
	c := NewLFUDA(10, nil)

	for i := 0; i < 100; i++ {
		c.Set(fmt.Sprintf("%v", i), i)
	}
	if c.Len() != 5 {
		t.Errorf("Failed to set or evict properly: %v", c.Len())
	}
}

func TestCacheFull(t *testing.T) {
	onEvicted := func(k interface{}, v interface{}) {
		if k == v {
			t.Errorf("Evict values equal (%v==%v) (but they shouldn't be)", k, v)
		}
	}

	c := NewLFUDA(3, onEvicted)
	c.Set("a", "a")
	if _, ok := c.Get("a"); !ok {
		t.Errorf("Key not found (but it should be)")
	}
	c.Set("b", "b")
	if _, ok := c.Get("b"); !ok {
		t.Errorf("Key not found (but it should be)")
	}
	if evict := c.Set("c", "z"); evict {
		t.Errorf("Set op resulted in an eviction (but it should not have)")
	}

	if evict := c.Set("d", "too big to store"); evict {
		t.Errorf("Set op resulted in an eviction (but it should not have)")
	}

	if evict := c.Set("d", "d"); !evict {
		t.Errorf("Set op did not result in an eviction (but it should have)")
	}

	if c.Age() != 1 {
		t.Errorf("Cache age should have incremented")
	}

	if _, ok := c.Get("a"); !ok {
		t.Errorf("Key not found (but it should be)")
	}

	if _, ok := c.Get("b"); !ok {
		t.Errorf("Key not found (but it should be)")
	}
}

func TestKeys(t *testing.T) {
	c := NewLFUDA(3, nil)
	c.Set("a", "a")
	c.Set("b", "b")
	c.Set("c", "c")

	if keys := c.Keys(); len(keys) != 3 || len(keys) != c.Len() {
		t.Errorf("Should be 3 keys returned from cache")
	}
}

func TestPurge(t *testing.T) {
	c := NewLFUDA(3, nil)
	c.Set("a", "a")
	c.Set("b", "b")
	c.Set("c", "c")

	if c.Len() != 3 {
		t.Errorf("Should be 3 keys in cache")
	}

	if !c.Contains("c") {
		t.Errorf("Cache should contain key c")
	}

	c.Purge()

	if c.Len() != 0 || c.Contains("c") {
		t.Errorf("Cache should be empty")
	}
}

func TestPeek(t *testing.T) {
	c := NewLFUDA(2, nil)
	c.Set("a", "a")
	c.Set("b", "b")

	// set key a to more frequent so b will be evicted
	if _, ok := c.Get("a"); !ok {
		t.Errorf("Key not found (but it should be)")
	}

	if val, _ := c.Peek("b"); val != "b" {
		t.Errorf("Key not found (but it should be)")
	}
	c.Peek("b")

	if evicted := c.Set("c", "c"); !evicted {
		t.Errorf("Set op should have resulted in eviction (but it did not)")
	}

	// b should be evicted
	if _, ok := c.Peek("b"); ok {
		t.Errorf("Key found (but it should not be)")
	}
}

func TestReSet(t *testing.T) {
	c := NewLFUDA(2, nil)
	c.Set("a", "a")
	if val, _ := c.Get("a"); val != "a" {
		t.Errorf("Keys not equal (but should be)")
	}

	// overwrite the key
	c.Set("a", "b")
	if val, _ := c.Get("a"); val != "b" {
		t.Errorf("Keys equal (but should not be)")
	}
}

func TestEvict(t *testing.T) {
	c := NewLFUDA(3, nil)
	c.Set("a", "a")
	c.Set("b", "b")
	c.Set("c", "c")

	// make key a popular
	for i := 0; i < 10; i++ {
		c.Get("a")
	}

	// increase cache age
	for j := 0; j < 2; j++ {
		for i := 0; i < 10; i++ {
			c.Set(i, i)
		}
	}

	if c.Age() != 10 {
		t.Errorf("cache should have aged for each eviction: %d", c.Age())
	}

	if ok := c.Contains("a"); !ok {
		t.Errorf("cache should have contained key a")
	}

	// kick out a
	for i := 0; i < 3; i++ {
		c.Set(i, i)
	}
	if ok := c.Contains("a"); ok {
		t.Errorf("cache should NOT have contained key a now")
	}
}

func TestEvictBigValue(t *testing.T) {
	c := NewLFUDA(10, nil)
	c.Set("a", "aaaaaaaa")
	c.Set("b", "b")
	c.Set("c", "c")

	if c.Size() != 10 {
		t.Errorf("cache should have size 10 bytes at this point: %d", c.Size())
	}

	// make key a popular
	for i := 0; i < 10; i++ {
		c.Get("a")
	}

	// increase cache age
	for j := 0; j < 2; j++ {
		for i := 0; i < 10; i++ {
			c.Set(i, i)
		}
	}

	if c.Age() != 10 {
		t.Errorf("cache should have aged for each eviction: %d", c.Age())
	}

	if ok := c.Contains("a"); !ok {
		t.Errorf("cache should have contained key a")
	}

	// may take couple to evict key a since it will share a frequency node with other keys at this point
	c.Set("x", "x")
	c.Set("y", "y")

	if ok := c.Contains("a"); ok {
		t.Errorf("cache should NOT have contained key a now")
	}

	if c.Size() > 4 || c.Size() < 3 {
		t.Errorf("cache should have size 3 or 4 bytes at this point: %d", c.Size())
	}
}
