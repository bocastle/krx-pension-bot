package cache

import (
	"sync"
	"time"
)

type item[V any] struct {
	value     V
	expiresAt time.Time
}

type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	ttl   time.Duration
	clock func() time.Time
	items map[K]item[V]
}

func New[K comparable, V any](ttl time.Duration, clock func() time.Time) *Cache[K, V] {
	if clock == nil {
		clock = time.Now
	}
	return &Cache[K, V]{
		ttl:   ttl,
		clock: clock,
		items: make(map[K]item[V]),
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()

	var zero V
	if !ok {
		return zero, false
	}
	if !c.clock().Before(entry.expiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return zero, false
	}
	return entry.value, true
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	c.items[key] = item[V]{
		value:     value,
		expiresAt: c.clock().Add(c.ttl),
	}
	c.mu.Unlock()
}
