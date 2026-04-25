package output

import (
	"bytes"
	"strings"
	"testing"
)

type testItem struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	items := []testItem{{Name: "foo", Value: 1}}
	err := Fprint(&buf, "json", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"name": "foo"`) {
		t.Fatalf("expected JSON output, got: %s", buf.String())
	}
}

func TestPrintYAML(t *testing.T) {
	var buf bytes.Buffer
	items := []testItem{{Name: "foo", Value: 1}}
	err := Fprint(&buf, "yaml", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "name: foo") {
		t.Fatalf("expected YAML output, got: %s", buf.String())
	}
}

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"NAME", "VALUE"}
	rows := [][]string{{"foo", "1"}, {"bar", "2"}}
	err := FprintTable(&buf, headers, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "foo") {
		t.Fatalf("expected table output, got: %s", out)
	}
}
