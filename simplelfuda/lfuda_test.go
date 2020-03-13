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
	c := NewLFUDA(10, nil)

	for i := 0; i < 100; i++ {
		c.Set(fmt.Sprintf("%v", i), i)
	}
	if c.Len() > 10 {
		t.Errorf("Failed to evict properly: %v", c.Len())
	}
}

func TestCacheFull(t *testing.T) {
	onEvicted := func(k interface{}, v interface{}) {
		if k == v {
			t.Errorf("Evict values equal (%v==%v) (but they shouldn't be)", k, v)
		}
	}

	c := NewLFUDA(3, onEvicted)
	if ok := c.Set("a", "a"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}
	if _, ok := c.Get("a"); !ok {
		t.Errorf("Key not found (but it should be)")
	}
	if ok := c.Set("b", "b"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}
	if _, ok := c.Get("b"); !ok {
		t.Errorf("Key not found (but it should be)")
	}
	if ok := c.Set("c", "evict"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}

	c.Set("d", "evict")
	c.Set("e", "evict")

	if _, ok := c.Get("a"); !ok {
		t.Errorf("Key not found (but it should be)")
	}

	if _, ok := c.Get("b"); !ok {
		t.Errorf("Key not found (but it should be)")
	}

	if _, ok := c.Get("c"); !ok {
		t.Errorf("Key not found (but it should be)")
	}
}

func TestKeys(t *testing.T) {
	c := NewLFUDA(3, nil)
	if ok := c.Set("a", "a"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}
	if ok := c.Set("b", "b"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}
	if ok := c.Set("c", "c"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}

	if keys := c.Keys(); len(keys) != 3 || len(keys) != c.Len() {
		t.Errorf("Should be 3 keys returned from cache")
	}
}

func TestPurge(t *testing.T) {
	c := NewLFUDA(3, nil)
	if ok := c.Set("a", "a"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}
	if ok := c.Set("b", "b"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}
	if ok := c.Set("c", "c"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}

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
	if ok := c.Set("a", "a"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}
	if ok := c.Set("b", "b"); !ok {
		t.Errorf("Set op was not successful (but it should have been)")
	}
	// set key a to more frequent so b will be evicted
	if _, ok := c.Get("a"); !ok {
		t.Errorf("Key not found (but it should be)")
	}

	if val, _ := c.Peek("b"); val != "b" {
		t.Errorf("Key not found (but it should be)")
	}
	c.Peek("b")

	if ok := c.Set("c", "c"); ok {
		t.Errorf("Set op successful (but it should NOT have been)")
	}

	// b should be evicted
	if _, ok := c.Peek("c"); ok {
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
	set := []string{"a", "b", "c"}

	for k := range set {
		if ok := c.Set(k, k); !ok {
			t.Errorf("Unable to set key %d", k)
		}
		if _, ok := c.Get(k); !ok {
			t.Errorf("Unable to get key %d", k)
		}
	}

	// should not be able to set a key since misses counter should equal 0
	if ok := c.Set("oops", "oops"); ok {
		t.Errorf("Able to set key (but should not have)")
	}

	// for misses counter
	for i := 0; i < 2; i++ {
		c.Get("missing")
	}

	// now SHOULD be able to set a key since misses counter should equal 2
	// and the least frequently used item's freq counter should also equal 2
	if ok := c.Set("huzzah", "huzzah"); !ok {
		t.Errorf("Not able to set key (but should have been)")
	}
}
