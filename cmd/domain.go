package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/swagner-de/desec-cli/internal/client"
	"github.com/swagner-de/desec-cli/internal/output"
)

var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage domains",
}

var domainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		domains, err := c.ListDomains()
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			headers := []string{"NAME", "CREATED", "PUBLISHED", "MIN TTL"}
			var rows [][]string
			for _, d := range domains {
				rows = append(rows, []string{
					d.Name,
					d.Created.Format("2006-01-02 15:04"),
					d.Published.Format("2006-01-02 15:04"),
					fmt.Sprintf("%d", d.MinimumTTL),
				})
			}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, domains)
	},
}

var domainGetCmd = &cobra.Command{
	Use:   "get <domain>",
	Short: "Get domain details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		domain, err := c.GetDomain(args[0])
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			headers := []string{"NAME", "CREATED", "PUBLISHED", "MIN TTL", "KEYS"}
			rows := [][]string{{
				domain.Name,
				domain.Created.Format("2006-01-02 15:04"),
				domain.Published.Format("2006-01-02 15:04"),
				fmt.Sprintf("%d", domain.MinimumTTL),
				fmt.Sprintf("%d", len(domain.Keys)),
			}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, domain)
	},
}

var domainCreateZonefile string

var domainCreateCmd = &cobra.Command{
	Use:   "create <domain>",
	Short: "Create a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		domain, err := c.CreateDomain(args[0], domainCreateZonefile)
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			fmt.Fprintf(os.Stderr, "Domain %s created.\n", domain.Name)
			headers := []string{"NAME", "CREATED", "MIN TTL"}
			rows := [][]string{{
				domain.Name,
				domain.Created.Format("2006-01-02 15:04"),
				fmt.Sprintf("%d", domain.MinimumTTL),
			}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, domain)
	},
}

var domainDeleteYes bool

var domainDeleteCmd = &cobra.Command{
	Use:   "delete <domain>",
	Short: "Delete a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !domainDeleteYes {
			fmt.Fprintf(os.Stderr, "Delete domain %s? This cannot be undone. [y/N]: ", args[0])
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(answer)) != "y" {
				fmt.Fprintln(os.Stderr, "Aborted.")
				return nil
			}
		}
		c := client.New(getToken())
		if err := c.DeleteDomain(args[0]); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Domain %s deleted.\n", args[0])
		return nil
	},
}

var domainExportCmd = &cobra.Command{
	Use:   "export <domain>",
	Short: "Export domain as zonefile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		zonefile, err := c.ExportDomain(args[0])
		if err != nil {
			return err
		}
		fmt.Print(zonefile)
		return nil
	},
}

func init() {
	domainCreateCmd.Flags().StringVar(&domainCreateZonefile, "zonefile", "", "Import records from zonefile content")
	domainDeleteCmd.Flags().BoolVarP(&domainDeleteYes, "yes", "y", false, "Skip confirmation prompt")

	domainCmd.AddCommand(domainListCmd)
	domainCmd.AddCommand(domainGetCmd)
	domainCmd.AddCommand(domainCreateCmd)
	domainCmd.AddCommand(domainDeleteCmd)
	domainCmd.AddCommand(domainExportCmd)
	rootCmd.AddCommand(domainCmd)
}
