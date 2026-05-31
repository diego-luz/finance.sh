// Package lockout implements a simple per-key failed-attempt counter used to
// throttle brute-force login attempts. State lives in process memory with a
// sliding TTL window — Community Edition runs a single API process so a shared
// store is unnecessary. All operations are nil-safe.
package lockout

import (
	"context"
	"sync"
	"time"
)

// Limiter tracks failed attempts per key with a sliding TTL window.
type Limiter struct {
	maxAttempts int
	ttl         time.Duration

	mu  sync.Mutex
	mem map[string]memEntry
}

type memEntry struct {
	count     int
	expiresAt time.Time
}

// New builds an in-memory Limiter.
func New(maxAttempts, lockoutMinutes int) *Limiter {
	if maxAttempts < 1 {
		maxAttempts = 5
	}
	if lockoutMinutes < 1 {
		lockoutMinutes = 15
	}
	return &Limiter{
		maxAttempts: maxAttempts,
		ttl:         time.Duration(lockoutMinutes) * time.Minute,
		mem:         make(map[string]memEntry),
	}
}

// MaxAttempts returns the configured failure threshold.
func (l *Limiter) MaxAttempts() int { return l.maxAttempts }

// LockoutMinutes returns the configured lockout window in minutes.
func (l *Limiter) LockoutMinutes() int { return int(l.ttl.Minutes()) }

// Locked reports whether the key currently exceeds the attempt threshold.
func (l *Limiter) Locked(_ context.Context, k string) bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.mem[k]
	if time.Now().After(e.expiresAt) {
		return false
	}
	return e.count >= l.maxAttempts
}

// RegisterFailure increments the failure counter for the key and (re)sets the
// TTL. It returns true if, after this failure, the key is now locked.
func (l *Limiter) RegisterFailure(_ context.Context, k string) bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.mem[k]
	if time.Now().After(e.expiresAt) {
		e = memEntry{}
	}
	e.count++
	e.expiresAt = time.Now().Add(l.ttl)
	l.mem[k] = e
	return e.count >= l.maxAttempts
}

// Reset clears the counter for the key (call on a successful login).
func (l *Limiter) Reset(_ context.Context, k string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	delete(l.mem, k)
	l.mu.Unlock()
}
