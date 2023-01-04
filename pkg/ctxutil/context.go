package ctxutil

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// NewCancelableContext creates a context that can be cancelled via a console termination, by starting a background
// goroutine listening to termination events.
func NewCancelableContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		select {
		case <-sig:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}
