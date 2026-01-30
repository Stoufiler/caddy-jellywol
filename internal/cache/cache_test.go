package cache

import (
	"testing"
	"time"
)

func TestNewResponseCache(t *testing.T) {
	cache := NewResponseCache(5 * time.Minute)
	if cache == nil {
		t.Fatal("Expected cache to be created")
	}

	if cache.Size() != 0 {
		t.Errorf("Expected cache size to be 0, got %d", cache.Size())
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewResponseCache(5 * time.Minute)

	key := "test-key"
	data := []byte("test data")
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}

	cache.Set(key, data, headers)

	entry, found := cache.Get(key)
	if !found {
		t.Fatal("Expected to find cached entry")
	}

	if string(entry.Data) != string(data) {
		t.Errorf("Expected data %s, got %s", data, entry.Data)
	}

	if entry.Headers["Content-Type"][0] != "application/json" {
		t.Errorf("Expected Content-Type header to be application/json")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := NewResponseCache(100 * time.Millisecond)

	key := "test-key"
	data := []byte("test data")

	cache.Set(key, data, nil)

	// Should exist immediately
	_, found := cache.Get(key)
	if !found {
		t.Fatal("Expected to find cached entry")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist after expiration
	_, found = cache.Get(key)
	if found {
		t.Fatal("Expected entry to be expired")
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewResponseCache(5 * time.Minute)

	key := "test-key"
	data := []byte("test data")

	cache.Set(key, data, nil)

	_, found := cache.Get(key)
	if !found {
		t.Fatal("Expected to find cached entry")
	}

	cache.Delete(key)

	_, found = cache.Get(key)
	if found {
		t.Fatal("Expected entry to be deleted")
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewResponseCache(5 * time.Minute)

	cache.Set("key1", []byte("data1"), nil)
	cache.Set("key2", []byte("data2"), nil)
	cache.Set("key3", []byte("data3"), nil)

	if cache.Size() != 3 {
		t.Errorf("Expected cache size to be 3, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected cache size to be 0 after clear, got %d", cache.Size())
	}
}

func TestCacheSize(t *testing.T) {
	cache := NewResponseCache(5 * time.Minute)

	if cache.Size() != 0 {
		t.Errorf("Expected initial cache size to be 0, got %d", cache.Size())
	}

	cache.Set("key1", []byte("data1"), nil)
	cache.Set("key2", []byte("data2"), nil)

	if cache.Size() != 2 {
		t.Errorf("Expected cache size to be 2, got %d", cache.Size())
	}
}
