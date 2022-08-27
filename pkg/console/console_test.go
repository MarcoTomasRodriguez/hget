package console

import (
	"context"
	"github.com/stretchr/testify/assert"
	"syscall"
	"testing"
	"time"
)

func TestCancelableContext_SIGINT(t *testing.T) {
	// Initialize empty cancelable context.
	ctx := CancelableContext(context.TODO())
	time.Sleep(10 * time.Millisecond)

	// Send kill signal.
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(10 * time.Millisecond)

	// Assert that context was cancelled.
	assert.ErrorIs(t, ctx.Err(), context.Canceled, "context should be canceled")
}

func TestCancelableContext_Cancel(t *testing.T) {
	// Initialize empty cancelable context.
	ctx, cancel := context.WithCancel(context.TODO())
	ctx = CancelableContext(ctx)
	time.Sleep(10 * time.Millisecond)

	// Cancel context.
	cancel()
	time.Sleep(10 * time.Millisecond)

	// Assert that context was cancelled.
	assert.ErrorIs(t, ctx.Err(), context.Canceled, "context should be canceled")
}
