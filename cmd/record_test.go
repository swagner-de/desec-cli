package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/swagner-de/desec-cli/internal/client"
)

func TestConfirmRecordDeletion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		yes   bool
		want  bool
	}{
		{name: "confirmed", input: "y\n", want: true},
		{name: "confirmed uppercase", input: "Y\n", want: true},
		{name: "aborted", input: "n\n", want: false},
		{name: "empty", input: "\n", want: false},
		{name: "yes flag", yes: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.SetIn(strings.NewReader(tt.input))
			cmd.SetErr(io.Discard)

			got, err := confirmRecordDeletion(cmd, "example.com", "www", "A", tt.yes)
			if err != nil {
				t.Fatalf("confirmRecordDeletion() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("confirmRecordDeletion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBulkResultMessage(t *testing.T) {
	if got := bulkResultMessage(nil); got != "Bulk update applied." {
		t.Fatalf("bulkResultMessage(nil) = %q", got)
	}
	if got := bulkResultMessage([]client.RRset{{}, {}}); got != "2 records processed." {
		t.Fatalf("bulkResultMessage() = %q", got)
	}
}
