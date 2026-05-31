// Package cache provides a tiny in-memory JSON cache used for the dashboard.
// All operations are best-effort: a nil *Cache is safe and turns every call
// into a no-op / miss. A single backend process owns the state — that fits
// finance.sh Community Edition (single-host, single API replica).
package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Cache holds JSON entries keyed by string with per-entry TTL. Concurrent
// access is safe.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]entry
}

type entry struct {
	value     []byte
	expiresAt time.Time
}

// New returns an in-memory Cache ready for use.
func New() *Cache {
	return &Cache{entries: make(map[string]entry)}
}

// GetJSON unmarshals the value at key into dst. Returns false on miss, expiry
// or any decode error (treated as a miss).
func (c *Cache) GetJSON(_ context.Context, key string, dst interface{}) bool {
	if c == nil {
		return false
	}
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().After(e.expiresAt) {
		// Lazy expiry: clear and report miss.
		c.mu.Lock()
		if cur, still := c.entries[key]; still && cur.expiresAt.Equal(e.expiresAt) {
			delete(c.entries, key)
		}
		c.mu.Unlock()
		return false
	}
	if err := json.Unmarshal(e.value, dst); err != nil {
		return false
	}
	return true
}

// SetJSON marshals v and stores it at key with the given TTL. Errors are
// swallowed (best-effort).
func (c *Cache) SetJSON(_ context.Context, key string, v interface{}, ttl time.Duration) {
	if c == nil {
		return
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return
	}
	c.mu.Lock()
	c.entries[key] = entry{value: raw, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

// Delete removes a key. Best-effort.
func (c *Cache) Delete(_ context.Context, key string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}
