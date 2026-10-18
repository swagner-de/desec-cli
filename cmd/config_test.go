package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestReadTokenFromPipe(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("  secret-token  \n"))
	cmd.SetErr(io.Discard)

	token, err := readToken(cmd)
	if err != nil {
		t.Fatalf("readToken() error = %v", err)
	}
	if token != "secret-token" {
		t.Fatalf("readToken() = %q, want %q", token, "secret-token")
	}
}

func TestReadTokenWithoutTrailingNewline(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("secret-token"))
	cmd.SetErr(io.Discard)

	token, err := readToken(cmd)
	if err != nil {
		t.Fatalf("readToken() error = %v", err)
	}
	if token != "secret-token" {
		t.Fatalf("readToken() = %q, want %q", token, "secret-token")
	}
}

func TestEnsureConfigWritable(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(path, []byte("token: old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ensureConfigWritable(path, false); err == nil {
		t.Fatal("expected existing config to require --force")
	}
	if err := ensureConfigWritable(path, true); err != nil {
		t.Fatalf("ensureConfigWritable(force) error = %v", err)
	}
	if err := ensureConfigWritable(t.TempDir()+"/missing.yaml", false); err != nil {
		t.Fatalf("ensureConfigWritable(missing) error = %v", err)
	}
}

func TestShouldPrintTokenSecret(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	if shouldPrintTokenSecret(cmd, false) {
		t.Fatal("expected secrets to be hidden when stderr is not a terminal")
	}
	if !shouldPrintTokenSecret(cmd, true) {
		t.Fatal("expected --show-secret to force secret output")
	}
}
