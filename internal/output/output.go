package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

func Print(format string, data any) error {
	return Fprint(os.Stdout, format, data)
}

func Fprint(w io.Writer, format string, data any) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "yaml":
		enc := yaml.NewEncoder(w)
		enc.SetIndent(2)
		defer enc.Close()
		return enc.Encode(data)
	default:
		return fmt.Errorf("unsupported format %q — use PrintTable for table output", format)
	}
}

func PrintTable(headers []string, rows [][]string) error {
	return FprintTable(os.Stdout, headers, rows)
}

func FprintTable(w io.Writer, headers []string, rows [][]string) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows)
	table.Render()
	return nil
}
