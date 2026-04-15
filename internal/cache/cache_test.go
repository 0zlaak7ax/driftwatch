package cache

import (
	"testing"
	"time"
)

func TestSet_And_Get_Valid(t *testing.T) {
	c := New(5 * time.Second)
	data := map[string]interface{}{"version": "1.2.3"}
	c.Set("svc-a", data)

	got, ok := c.Get("svc-a")
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if got["version"] != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %v", got["version"])
	}
}

func TestGet_Miss(t *testing.T) {
	c := New(5 * time.Second)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss, got hit")
	}
}

func TestGet_Expired(t *testing.T) {
	c := New(10 * time.Millisecond)
	c.Set("svc-b", map[string]interface{}{"status": "ok"})

	time.Sleep(20 * time.Millisecond)

	_, ok := c.Get("svc-b")
	if ok {
		t.Fatal("expected expired entry to be a cache miss")
	}
}

func TestInvalidate(t *testing.T) {
	c := New(5 * time.Second)
	c.Set("svc-c", map[string]interface{}{"env": "prod"})
	c.Invalidate("svc-c")

	_, ok := c.Get("svc-c")
	if ok {
		t.Fatal("expected invalidated entry to be a cache miss")
	}
}

func TestPurge_RemovesExpired(t *testing.T) {
	c := New(10 * time.Millisecond)
	c.Set("svc-d", map[string]interface{}{"x": 1})
	c.Set("svc-e", map[string]interface{}{"x": 2})

	time.Sleep(20 * time.Millisecond)
	c.Purge()

	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.entries) != 0 {
		t.Errorf("expected 0 entries after purge, got %d", len(c.entries))
	}
}

func TestPurge_KeepsValid(t *testing.T) {
	c := New(5 * time.Second)
	c.Set("svc-f", map[string]interface{}{"alive": true})
	c.Purge()

	_, ok := c.Get("svc-f")
	if !ok {
		t.Fatal("expected valid entry to survive purge")
	}
}
