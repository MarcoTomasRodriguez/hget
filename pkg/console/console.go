package console

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// CancelableContext creates a context that can be canceled from the console.
// For example, by typing Ctrl + C.
func CancelableContext(ctx context.Context) context.Context {
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
