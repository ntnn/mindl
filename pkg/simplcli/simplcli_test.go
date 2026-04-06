package simplcli

import (
	"bytes"
	"context"
	"errors"
	"os"
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
			runner := func(_ context.Context, args []string) error {
				capturedArgs = args
				return nil
			}
			cli := SimplCLI{
				SubCmds: map[string]SubCmd{
					"greet": {Runner: runner, Doc: "say hello"},
				},
			}

			err := cli.Run(context.Background(), tc.args)
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
				Runner: func(_ context.Context, _ []string) error {
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

			err := cli.Run(context.Background(), tc.args)
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

//nolint:paralleltest // test captures os.Stdout which is global state
func TestSimplCLIRunHelp(t *testing.T) {
	// "help" as first arg should not return an error
	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"greet": {
				Runner: func(_ context.Context, _ []string) error { return nil },
				Doc:    "say hello",
			},
		},
	}

	// Capture stdout since PrintHelp writes there
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cli.Run(context.Background(), []string{"help"})

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}
	os.Stdout = old

	if err != nil {
		t.Fatalf("Run(help) returned unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
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
				Runner: func(_ context.Context, _ []string) error {
					customHelpCalled = true
					return nil
				},
				Doc: "custom help",
			},
			"greet": {
				Runner: func(_ context.Context, _ []string) error { return nil },
				Doc:    "say hello",
			},
		},
	}

	err := cli.Run(context.Background(), []string{"help"})
	if err != nil {
		t.Fatalf("Run(help) returned unexpected error: %v", err)
	}
	if !customHelpCalled {
		t.Error("expected custom help runner to be called")
	}
}

//nolint:paralleltest // test captures os.Stdout which is global state
func TestPrintDefaultHelp(t *testing.T) {
	cli := SimplCLI{
		SubCmds: map[string]SubCmd{
			"beta":  {Doc: "beta command"},
			"alpha": {Doc: "alpha command"},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintDefaultHelp(cli)

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
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
