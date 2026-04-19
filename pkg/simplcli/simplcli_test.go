package simplcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestSimplCLIRun(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args     []string
		wantArgs []string
	}{
		"subcommand with no extra args": {
			args:     []string{"greet"},
			wantArgs: []string{},
		},
		"subcommand with args": {
			args:     []string{"greet", "world", "foo"},
			wantArgs: []string{"world", "foo"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var capturedArgs []string
			runner := func(_ context.Context, _, _ io.Writer, args []string) error {
				capturedArgs = args
				return nil
			}
			cli := SimplCLI{
				SubCmds: map[string]SubCmd{
					"greet": {Runner: runner, Doc: "say hello"},
				},
			}

			err := cli.Run(t.Context(), io.Discard, io.Discard, tc.args)
			if err != nil {
				t.Fatalf("Run returned unexpected error: %v", err)
			}
			if len(capturedArgs) != len(tc.wantArgs) {
				t.Fatalf("runner received %d args, want %d", len(capturedArgs), len(tc.wantArgs))
			}
			for i, want := range tc.wantArgs {
				if capturedArgs[i] != want {
					t.Errorf("arg[%d] = %q, want %q", i, capturedArgs[i], want)
				}
			}
		})
	}
}

func TestSimplCLIRunError(t *testing.T) {
	t.Parallel()

	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"fail": {
				Runner: func(_ context.Context, _, _ io.Writer, _ []string) error {
					return errors.New("runner failed")
				},
				Doc: "always fails",
			},
		},
	}

	tests := map[string]struct {
		args       []string
		wantErr    error
		wantSubstr string
	}{
		"no args": {
			args:    []string{},
			wantErr: ErrNoArgs,
		},
		"unknown subcommand": {
			args:       []string{"nonexistent"},
			wantSubstr: "unknown subcommand",
		},
		"runner returns error": {
			args:       []string{"fail"},
			wantSubstr: "runner failed",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := cli.Run(t.Context(), io.Discard, io.Discard, tc.args)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Errorf("error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantSubstr != "" && !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Errorf("error = %q, want it to contain %q", err, tc.wantSubstr)
			}
		})
	}
}

func TestSimplCLIRunHelp(t *testing.T) {
	t.Parallel()

	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"greet": {
				Runner: func(_ context.Context, _, _ io.Writer, _ []string) error { return nil },
				Doc:    "say hello",
			},
		},
	}

	var buf bytes.Buffer
	err := cli.Run(t.Context(), &buf, io.Discard, []string{"help"})
	if err != nil {
		t.Fatalf("Run(help) returned unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "greet") {
		t.Errorf("help output should contain subcommand name, got %q", output)
	}
}

func TestSimplCLIRunCustomHelp(t *testing.T) {
	t.Parallel()

	var customHelpCalled bool
	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"help": {
				Runner: func(_ context.Context, _, _ io.Writer, _ []string) error {
					customHelpCalled = true
					return nil
				},
				Doc: "custom help",
			},
			"greet": {
				Runner: func(_ context.Context, _, _ io.Writer, _ []string) error { return nil },
				Doc:    "say hello",
			},
		},
	}

	err := cli.Run(t.Context(), io.Discard, io.Discard, []string{"help"})
	if err != nil {
		t.Fatalf("Run(help) returned unexpected error: %v", err)
	}
	if !customHelpCalled {
		t.Error("expected custom help runner to be called")
	}
}

func TestPrintDefaultHelp(t *testing.T) {
	t.Parallel()

	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"beta":  {Doc: "beta command"},
			"alpha": {Doc: "alpha command"},
		},
	}

	var buf bytes.Buffer
	err := PrintDefaultHelp(&buf, cli)
	if err != nil {
		t.Fatalf("unexpected error printint help: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Available subcommands:") {
		t.Errorf("output should contain header, got %q", output)
	}

	// Verify sorted order: alpha before beta
	alphaIdx := strings.Index(output, "alpha")
	betaIdx := strings.Index(output, "beta")
	if alphaIdx == -1 || betaIdx == -1 {
		t.Fatalf("output should contain both 'alpha' and 'beta', got %q", output)
	}
	if alphaIdx > betaIdx {
		t.Errorf("expected alpha before beta in sorted output")
	}

	// Verify docs are present
	if !strings.Contains(output, "alpha command") {
		t.Errorf("output should contain 'alpha command', got %q", output)
	}
	if !strings.Contains(output, "beta command") {
		t.Errorf("output should contain 'beta command', got %q", output)
	}
}

func TestSimplCLIMiddlewareExecutionOrder(t *testing.T) {
	t.Parallel()

	var order []string
	mw1 := func(ctx context.Context, stdout, stderr io.Writer, args []string, next Runner) error {
		order = append(order, "mw1")
		return next(ctx, stdout, stderr, args)
	}
	mw2 := func(ctx context.Context, stdout, stderr io.Writer, args []string, next Runner) error {
		order = append(order, "mw2")
		return next(ctx, stdout, stderr, args)
	}
	runner := func(_ context.Context, _, _ io.Writer, _ []string) error {
		order = append(order, "runner")
		return nil
	}
	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"test": {Runner: runner},
		},
		Middlewares: []Middleware{mw1, mw2},
	}

	if err := cli.Run(t.Context(), io.Discard, io.Discard, []string{"test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"mw1", "mw2", "runner"}
	if len(order) != len(expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}
	for i, want := range expected {
		if order[i] != want {
			t.Errorf("at index %d, expected %q, got %q", i, want, order[i])
		}
	}
}

func TestSimplCLIMiddlewareInterceptAndStop(t *testing.T) {
	t.Parallel()
	mw := func(_ context.Context, _, _ io.Writer, _ []string, _ Runner) error {
		return errors.New("middleware blocked")
	}
	runnerCalled := false
	runner := func(_ context.Context, _, _ io.Writer, _ []string) error {
		runnerCalled = true
		return nil
	}
	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"test": {Runner: runner},
		},
		Middlewares: []Middleware{mw},
	}

	err := cli.Run(t.Context(), io.Discard, io.Discard, []string{"test"})
	if err == nil || err.Error() != "middleware blocked" {
		t.Fatalf("expected error 'middleware blocked', got %v", err)
	}
	if runnerCalled {
		t.Error("runner should not have been called")
	}
}

func TestSimplCLIMiddlewareModifyStdout(t *testing.T) {
	t.Parallel()
	mw := func(ctx context.Context, stdout, stderr io.Writer, args []string, next Runner) error {
		_, _ = fmt.Fprint(stdout, "mw-prefix ")
		return next(ctx, stdout, stderr, args)
	}
	runner := func(_ context.Context, stdout, _ io.Writer, _ []string) error {
		_, _ = fmt.Fprint(stdout, "runner-output")
		return nil
	}
	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"test": {Runner: runner},
		},
		Middlewares: []Middleware{mw},
	}

	var buf bytes.Buffer
	if err := cli.Run(t.Context(), &buf, io.Discard, []string{"test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "mw-prefix runner-output"
	if buf.String() != want {
		t.Errorf("expected output %q, got %q", want, buf.String())
	}
}

func TestSimplCLIMiddlewareModifyContext(t *testing.T) {
	t.Parallel()
	type ctxKey string
	const key ctxKey = "mykey"
	const val = "myval"

	mw := func(ctx context.Context, stdout, stderr io.Writer, args []string, next Runner) error {
		ctx = context.WithValue(ctx, key, val)
		return next(ctx, stdout, stderr, args)
	}
	runner := func(ctx context.Context, _, _ io.Writer, _ []string) error {
		if ctx.Value(key) != val {
			return errors.New("context value not found")
		}
		return nil
	}
	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"test": {Runner: runner},
		},
		Middlewares: []Middleware{mw},
	}

	if err := cli.Run(t.Context(), io.Discard, io.Discard, []string{"test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
