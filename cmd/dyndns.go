package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/swagner-de/desec-cli/internal/client"
)

var dyndnsCmd = &cobra.Command{
	Use:   "dyndns",
	Short: "Dynamic DNS operations",
}

var dyndnsHostname string
var dyndnsIPv4 string
var dyndnsIPv6 string

var dyndnsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update dynamic DNS IP address",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dyndnsIPv4 == "" && dyndnsIPv6 == "" {
			return fmt.Errorf("at least one of --ipv4 or --ipv6 must be specified")
		}
		c := client.New(getToken())
		if err := c.DynDNSUpdate(dyndnsHostname, dyndnsIPv4, dyndnsIPv6); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "DNS updated.")
		return nil
	},
}

func init() {
	dyndnsUpdateCmd.Flags().StringVar(&dyndnsHostname, "hostname", "", "Hostname to update")
	dyndnsUpdateCmd.Flags().StringVar(&dyndnsIPv4, "ipv4", "", "IPv4 address")
	dyndnsUpdateCmd.Flags().StringVar(&dyndnsIPv6, "ipv6", "", "IPv6 address")
	dyndnsCmd.AddCommand(dyndnsUpdateCmd)
	rootCmd.AddCommand(dyndnsCmd)
}
