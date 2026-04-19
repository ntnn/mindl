package simplcli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// WithSignal wraps a [context.Context] to be cancelled when the specified signals are received.
// This is virtually equivalent to [signal.NotifyContext] but prints a message to stderr.
// When no signals are passed os.Interrupt and syscal.SIGTERM are defaulted.
func WithSignal(ctx context.Context, stderr io.Writer, signals ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx) //nolint:gosec // cancel function is returned to caller

	if signals == nil {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), signals...)

	go func() {
		<-signalCtx.Done()
		if ctx.Err() == nil {
			return
		}
		if strings.Contains(ctx.Err().Error(), "signal.NotifyContext") {
			_, _ = fmt.Fprintln(stderr, "Received signal to shut down...")
			cancel()
		}
	}()

	return ctx, func() { cancel(); stop() }
}

// Entrypoint runs [SimplCLI] with a signal-cancelled [context.Context], [os.Stdout], [os.Stderr] and [os.Args].
// If an error is returned it is printed to [os.Stderr] and [os.Exit] is called.
func Entrypoint(simplcli SimplCLI) {
	ctx, cancel := WithSignal(context.Background(), os.Stderr)

	err := simplcli.Run(ctx, os.Stdout, os.Stderr, os.Args[1:])
	cancel()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
