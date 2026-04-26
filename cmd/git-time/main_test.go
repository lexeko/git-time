package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	tests := [][]string{{"--help"}, {"-h"}}

	for _, args := range tests {
		t.Run(args[0], func(t *testing.T) {
			var stdout bytes.Buffer
			if err := run(args, &stdout); err != nil {
				t.Fatalf("run returned error: %v", err)
			}

			out := stdout.String()
			if !strings.Contains(out, "Usage:") {
				t.Fatalf("help output missing Usage: %q", out)
			}
			if !strings.Contains(out, "--help, -h") {
				t.Fatalf("help output missing help option: %q", out)
			}
		})
	}
}
