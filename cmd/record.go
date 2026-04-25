package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/swagner-de/desec-cli/internal/client"
	"github.com/swagner-de/desec-cli/internal/output"
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Manage DNS records (RRsets)",
}

var recordListType string
var recordListSubname string

var recordListCmd = &cobra.Command{
	Use:   "list <domain>",
	Short: "List DNS records",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		rrsets, err := c.ListRRsets(args[0], recordListType, recordListSubname)
		if err != nil { return err }
		if outputFormat == "table" {
			headers := []string{"SUBNAME", "TYPE", "TTL", "RECORDS"}
			var rows [][]string
			for _, r := range rrsets {
				rows = append(rows, []string{displaySubname(r.Subname), r.Type, fmt.Sprintf("%d", r.TTL), strings.Join(r.Records, ", ")})
			}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, rrsets)
	},
}

var recordGetCmd = &cobra.Command{
	Use:   "get <domain> <subname> <type>",
	Short: "Get a specific DNS record",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		subname := apiSubname(args[1])
		rrset, err := c.GetRRset(args[0], subname, args[2])
		if err != nil { return err }
		if outputFormat == "table" {
			headers := []string{"SUBNAME", "TYPE", "TTL", "RECORDS"}
			rows := [][]string{{displaySubname(rrset.Subname), rrset.Type, fmt.Sprintf("%d", rrset.TTL), strings.Join(rrset.Records, ", ")}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, rrset)
	},
}

var recordCreateSubname string
var recordCreateType string
var recordCreateTTL int
var recordCreateRecords []string

var recordCreateCmd = &cobra.Command{
	Use:   "create <domain>",
	Short: "Create a DNS record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		rrset, err := c.CreateRRset(args[0], &client.RRsetCreate{
			Subname: apiSubname(recordCreateSubname),
			Type:    recordCreateType,
			Records: recordCreateRecords,
			TTL:     recordCreateTTL,
		})
		if err != nil { return err }
		if outputFormat == "table" {
			fmt.Fprintln(os.Stderr, "Record created.")
			headers := []string{"SUBNAME", "TYPE", "TTL", "RECORDS"}
			rows := [][]string{{displaySubname(rrset.Subname), rrset.Type, fmt.Sprintf("%d", rrset.TTL), strings.Join(rrset.Records, ", ")}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, rrset)
	},
}

var recordUpdateTTL int
var recordUpdateRecords []string

var recordUpdateCmd = &cobra.Command{
	Use:   "update <domain> <subname> <type>",
	Short: "Update a DNS record",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		update := map[string]any{}
		if cmd.Flags().Changed("ttl") { update["ttl"] = recordUpdateTTL }
		if cmd.Flags().Changed("record") { update["records"] = recordUpdateRecords }
		if len(update) == 0 { return fmt.Errorf("nothing to update — specify --ttl and/or --record") }
		rrset, err := c.UpdateRRset(args[0], apiSubname(args[1]), args[2], update)
		if err != nil { return err }
		if outputFormat == "table" {
			fmt.Fprintln(os.Stderr, "Record updated.")
			headers := []string{"SUBNAME", "TYPE", "TTL", "RECORDS"}
			rows := [][]string{{displaySubname(rrset.Subname), rrset.Type, fmt.Sprintf("%d", rrset.TTL), strings.Join(rrset.Records, ", ")}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, rrset)
	},
}

var recordDeleteCmd = &cobra.Command{
	Use:   "delete <domain> <subname> <type>",
	Short: "Delete a DNS record",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		if err := c.DeleteRRset(args[0], apiSubname(args[1]), args[2]); err != nil { return err }
		fmt.Fprintln(os.Stderr, "Record deleted.")
		return nil
	},
}

var recordBulkFile string

var recordBulkCmd = &cobra.Command{
	Use:   "bulk <domain>",
	Short: "Bulk create/update DNS records from JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var input io.Reader
		if recordBulkFile != "" {
			f, err := os.Open(recordBulkFile)
			if err != nil { return fmt.Errorf("opening file: %w", err) }
			defer f.Close()
			input = f
		} else {
			input = os.Stdin
		}
		data, err := io.ReadAll(input)
		if err != nil { return fmt.Errorf("reading input: %w", err) }
		var rrsets []client.RRsetCreate
		if err := json.Unmarshal(data, &rrsets); err != nil { return fmt.Errorf("parsing JSON: %w", err) }
		c := client.New(getToken())
		results, err := c.BulkRRsets(args[0], rrsets)
		if err != nil { return err }
		if outputFormat == "table" {
			fmt.Fprintf(os.Stderr, "%d records processed.\n", len(results))
			headers := []string{"SUBNAME", "TYPE", "TTL", "RECORDS"}
			var rows [][]string
			for _, r := range results { rows = append(rows, []string{displaySubname(r.Subname), r.Type, fmt.Sprintf("%d", r.TTL), strings.Join(r.Records, ", ")}) }
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, results)
	},
}

func apiSubname(s string) string {
	if s == "@" { return "" }
	return s
}

func displaySubname(s string) string {
	if s == "" { return "@" }
	return s
}

func init() {
	recordListCmd.Flags().StringVar(&recordListType, "type", "", "Filter by record type")
	recordListCmd.Flags().StringVar(&recordListSubname, "subname", "", "Filter by subname")
	recordCreateCmd.Flags().StringVar(&recordCreateSubname, "subname", "@", "Record subname (@ for zone apex)")
	recordCreateCmd.Flags().StringVar(&recordCreateType, "type", "", "Record type (A, AAAA, CNAME, MX, etc.)")
	recordCreateCmd.Flags().IntVar(&recordCreateTTL, "ttl", 3600, "TTL in seconds")
	recordCreateCmd.Flags().StringArrayVar(&recordCreateRecords, "record", nil, "Record value (repeatable)")
	_ = recordCreateCmd.MarkFlagRequired("type")
	_ = recordCreateCmd.MarkFlagRequired("record")
	recordUpdateCmd.Flags().IntVar(&recordUpdateTTL, "ttl", 0, "New TTL in seconds")
	recordUpdateCmd.Flags().StringArrayVar(&recordUpdateRecords, "record", nil, "New record value (repeatable, replaces all)")
	recordBulkCmd.Flags().StringVar(&recordBulkFile, "file", "", "Read JSON from file instead of stdin")
	recordCmd.AddCommand(recordListCmd)
	recordCmd.AddCommand(recordGetCmd)
	recordCmd.AddCommand(recordCreateCmd)
	recordCmd.AddCommand(recordUpdateCmd)
	recordCmd.AddCommand(recordDeleteCmd)
	recordCmd.AddCommand(recordBulkCmd)
	rootCmd.AddCommand(recordCmd)
}
