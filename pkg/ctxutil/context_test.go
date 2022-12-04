package ctxutil

import (
	"context"
	"github.com/stretchr/testify/assert"
	"syscall"
	"testing"
	"time"
)

func TestNewCancelableContext_SIGINT(t *testing.T) {
	// Initialize empty cancelable ctxutil.
	ctx := NewCancelableContext(context.TODO())
	time.Sleep(10 * time.Millisecond)

	// Send kill signal.
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(10 * time.Millisecond)

	// Assert that ctxutil was cancelled.
	assert.ErrorIs(t, ctx.Err(), context.Canceled, "ctxutil should be canceled")
}

func TestNewCancelableContext_Cancel(t *testing.T) {
	// Initialize empty cancelable ctxutil.
	ctx, cancel := context.WithCancel(context.TODO())
	ctx = NewCancelableContext(ctx)
	time.Sleep(10 * time.Millisecond)

	// Cancel ctxutil.
	cancel()
	time.Sleep(10 * time.Millisecond)

	// Assert that ctxutil was cancelled.
	assert.ErrorIs(t, ctx.Err(), context.Canceled, "ctxutil should be canceled")
}
