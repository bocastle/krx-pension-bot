package cache

import (
	"testing"
	"time"
)

func TestTTLCacheReturnsValueUntilExpired(t *testing.T) {
	now := time.Date(2026, 5, 26, 9, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	c := New[string, int](10*time.Minute, clock)

	c.Set("samsung", 7)

	got, ok := c.Get("samsung")
	if !ok || got != 7 {
		t.Fatalf("Get before expiry = %v, %v; want 7, true", got, ok)
	}

	now = now.Add(11 * time.Minute)

	got, ok = c.Get("samsung")
	if ok {
		t.Fatalf("Get after expiry = %v, true; want false", got)
	}
}
