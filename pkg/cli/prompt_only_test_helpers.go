package cli

import (
	"os"
	"testing"
)

func withStdin(t *testing.T, input string, fn func() string) string {
	t.Helper()

	previous := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	if _, err := writer.WriteString(input); err != nil {
		t.Fatalf("writer.WriteString() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	os.Stdin = reader
	defer func() {
		os.Stdin = previous
		_ = reader.Close()
	}()

	return fn()
}
