package cmd

import (
	"io"
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
