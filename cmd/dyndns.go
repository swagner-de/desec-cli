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
		if err := validateDynDNSUpdate(dyndnsHostname, dyndnsIPv4, dyndnsIPv6); err != nil {
			return err
		}
		c := client.New(getToken())
		if err := c.DynDNSUpdate(dyndnsHostname, dyndnsIPv4, dyndnsIPv6); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "DNS updated.")
		return nil
	},
}

func validateDynDNSUpdate(hostname, ipv4, ipv6 string) error {
	if hostname == "" {
		return fmt.Errorf("--hostname is required")
	}
	if ipv4 == "" || ipv6 == "" {
		return fmt.Errorf("both --ipv4 and --ipv6 are required; use \"preserve\" to keep an existing address")
	}
	return nil
}

func init() {
	dyndnsUpdateCmd.Flags().StringVar(&dyndnsHostname, "hostname", "", "Hostname to update")
	dyndnsUpdateCmd.Flags().StringVar(&dyndnsIPv4, "ipv4", "", "IPv4 address or \"preserve\"")
	dyndnsUpdateCmd.Flags().StringVar(&dyndnsIPv6, "ipv6", "", "IPv6 address or \"preserve\"")
	dyndnsCmd.AddCommand(dyndnsUpdateCmd)
	rootCmd.AddCommand(dyndnsCmd)
}
