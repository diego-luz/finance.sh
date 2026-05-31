package lockout_test

import (
	"context"
	"testing"

	"github.com/finance-sh/finance-sh/pkg/lockout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaults(t *testing.T) {
	// Out-of-range values fall back to defaults.
	l := lockout.New(0, 0)
	assert.Equal(t, 5, l.MaxAttempts())
	assert.Equal(t, 15, l.LockoutMinutes())

	l2 := lockout.New(3, 10)
	assert.Equal(t, 3, l2.MaxAttempts())
	assert.Equal(t, 10, l2.LockoutMinutes())
}

func TestLocksAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	const max = 3
	l := lockout.New(max, 15)

	const email = "user@finance.sh"
	assert.False(t, l.Locked(ctx, email), "fresh key is not locked")

	// First (max-1) failures do not lock yet.
	for i := 0; i < max-1; i++ {
		locked := l.RegisterFailure(ctx, email)
		assert.False(t, locked, "should not be locked before reaching the threshold (i=%d)", i)
		assert.False(t, l.Locked(ctx, email))
	}

	// The max-th failure locks.
	locked := l.RegisterFailure(ctx, email)
	assert.True(t, locked, "the threshold failure must report locked")
	assert.True(t, l.Locked(ctx, email))

	// Further failures keep it locked.
	assert.True(t, l.RegisterFailure(ctx, email))
	assert.True(t, l.Locked(ctx, email))
}

func TestResetClearsCounter(t *testing.T) {
	ctx := context.Background()
	l := lockout.New(2, 15)
	const email = "reset2@finance.sh"

	assert.False(t, l.RegisterFailure(ctx, email)) // count 1, max 2 -> not locked
	assert.True(t, l.RegisterFailure(ctx, email))  // count 2 -> locked
	assert.True(t, l.Locked(ctx, email))

	l.Reset(ctx, email)
	assert.False(t, l.Locked(ctx, email), "reset must clear the lock")

	// Counter restarts from zero after reset.
	assert.False(t, l.RegisterFailure(ctx, email))
	assert.True(t, l.RegisterFailure(ctx, email))
}

func TestEmailsAreIndependent(t *testing.T) {
	ctx := context.Background()
	l := lockout.New(2, 15)

	const a = "a@finance.sh"
	const b = "b@finance.sh"

	l.RegisterFailure(ctx, a)
	require.True(t, l.RegisterFailure(ctx, a)) // a now locked
	assert.True(t, l.Locked(ctx, a))

	assert.False(t, l.Locked(ctx, b), "b must be unaffected by a's failures")
	assert.False(t, l.RegisterFailure(ctx, b), "b's first failure is independent")
}

func TestNilLimiterIsSafe(t *testing.T) {
	ctx := context.Background()
	var l *lockout.Limiter // nil

	assert.False(t, l.Locked(ctx, "x@finance.sh"))
	assert.False(t, l.RegisterFailure(ctx, "x@finance.sh"))
	assert.NotPanics(t, func() { l.Reset(ctx, "x@finance.sh") })
}
