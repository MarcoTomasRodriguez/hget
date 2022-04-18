package console

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// CancelableContext creates a context that can be canceled from the console.
// For example, by typing Ctrl + C.
func CancelableContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
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
