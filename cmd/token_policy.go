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

func ptrStr(s *string) string {
	if s == nil {
		return "*"
	}
	return *s
}

var tokenPolicyCmd = &cobra.Command{
	Use:   "token-policy",
	Short: "Manage token policies",
}

var tokenPolicyListCmd = &cobra.Command{
	Use:   "list <token-id>",
	Short: "List policies for a token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		policies, err := c.ListPolicies(args[0])
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			headers := []string{"ID", "DOMAIN", "SUBNAME", "TYPE", "WRITE"}
			var rows [][]string
			for _, p := range policies {
				rows = append(rows, []string{
					p.ID,
					ptrStr(p.Domain),
					ptrStr(p.Subname),
					ptrStr(p.Type),
					fmt.Sprintf("%v", p.PermWrite),
				})
			}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, policies)
	},
}

var tokenPolicyGetCmd = &cobra.Command{
	Use:   "get <token-id> <policy-id>",
	Short: "Get a token policy",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		p, err := c.GetPolicy(args[0], args[1])
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			headers := []string{"ID", "DOMAIN", "SUBNAME", "TYPE", "WRITE"}
			rows := [][]string{{
				p.ID,
				ptrStr(p.Domain),
				ptrStr(p.Subname),
				ptrStr(p.Type),
				fmt.Sprintf("%v", p.PermWrite),
			}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, p)
	},
}

var tokenPolicyCreateDomain string
var tokenPolicyCreateSubname string
var tokenPolicyCreateType string
var tokenPolicyCreatePermWrite bool

var tokenPolicyCreateCmd = &cobra.Command{
	Use:   "create <token-id>",
	Short: "Create a token policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		pc := &client.TokenPolicyCreate{
			PermWrite: tokenPolicyCreatePermWrite,
		}
		if cmd.Flags().Changed("domain") {
			v := tokenPolicyCreateDomain
			pc.Domain = &v
		}
		if cmd.Flags().Changed("subname") {
			v := tokenPolicyCreateSubname
			pc.Subname = &v
		}
		if cmd.Flags().Changed("type") {
			v := tokenPolicyCreateType
			pc.Type = &v
		}
		p, err := c.CreatePolicy(args[0], pc)
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			fmt.Fprintln(os.Stderr, "Policy created.")
			headers := []string{"ID", "DOMAIN", "SUBNAME", "TYPE", "WRITE"}
			rows := [][]string{{
				p.ID,
				ptrStr(p.Domain),
				ptrStr(p.Subname),
				ptrStr(p.Type),
				fmt.Sprintf("%v", p.PermWrite),
			}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, p)
	},
}

var tokenPolicyUpdateDomain string
var tokenPolicyUpdateSubname string
var tokenPolicyUpdateType string
var tokenPolicyUpdatePermWrite bool

var tokenPolicyUpdateCmd = &cobra.Command{
	Use:   "update <token-id> <policy-id>",
	Short: "Update a token policy",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		update := map[string]any{}
		if cmd.Flags().Changed("domain") {
			v := tokenPolicyUpdateDomain
			update["domain"] = &v
		}
		if cmd.Flags().Changed("subname") {
			v := tokenPolicyUpdateSubname
			update["subname"] = &v
		}
		if cmd.Flags().Changed("type") {
			v := tokenPolicyUpdateType
			update["type"] = &v
		}
		if cmd.Flags().Changed("perm-write") {
			update["perm_write"] = tokenPolicyUpdatePermWrite
		}
		if len(update) == 0 {
			return fmt.Errorf("nothing to update — specify at least one flag")
		}
		p, err := c.UpdatePolicy(args[0], args[1], update)
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			fmt.Fprintln(os.Stderr, "Policy updated.")
			headers := []string{"ID", "DOMAIN", "SUBNAME", "TYPE", "WRITE"}
			rows := [][]string{{
				p.ID,
				ptrStr(p.Domain),
				ptrStr(p.Subname),
				ptrStr(p.Type),
				fmt.Sprintf("%v", p.PermWrite),
			}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, p)
	},
}

var tokenPolicyDeleteYes bool

var tokenPolicyDeleteCmd = &cobra.Command{
	Use:   "delete <token-id> <policy-id>",
	Short: "Delete a token policy",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tokenPolicyDeleteYes {
			fmt.Fprintf(os.Stderr, "Delete policy %s for token %s? [y/N]: ", args[1], args[0])
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(answer)) != "y" {
				fmt.Fprintln(os.Stderr, "Aborted.")
				return nil
			}
		}
		c := client.New(getToken())
		if err := c.DeletePolicy(args[0], args[1]); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Policy %s deleted.\n", args[1])
		return nil
	},
}

func init() {
	tokenPolicyCreateCmd.Flags().StringVar(&tokenPolicyCreateDomain, "domain", "", "Domain (omit for wildcard)")
	tokenPolicyCreateCmd.Flags().StringVar(&tokenPolicyCreateSubname, "subname", "", "Subname (omit for wildcard)")
	tokenPolicyCreateCmd.Flags().StringVar(&tokenPolicyCreateType, "type", "", "Record type (omit for wildcard)")
	tokenPolicyCreateCmd.Flags().BoolVar(&tokenPolicyCreatePermWrite, "perm-write", false, "Allow write access")

	tokenPolicyUpdateCmd.Flags().StringVar(&tokenPolicyUpdateDomain, "domain", "", "Domain")
	tokenPolicyUpdateCmd.Flags().StringVar(&tokenPolicyUpdateSubname, "subname", "", "Subname")
	tokenPolicyUpdateCmd.Flags().StringVar(&tokenPolicyUpdateType, "type", "", "Record type")
	tokenPolicyUpdateCmd.Flags().BoolVar(&tokenPolicyUpdatePermWrite, "perm-write", false, "Allow write access")

	tokenPolicyDeleteCmd.Flags().BoolVarP(&tokenPolicyDeleteYes, "yes", "y", false, "Skip confirmation prompt")

	tokenPolicyCmd.AddCommand(tokenPolicyListCmd)
	tokenPolicyCmd.AddCommand(tokenPolicyGetCmd)
	tokenPolicyCmd.AddCommand(tokenPolicyCreateCmd)
	tokenPolicyCmd.AddCommand(tokenPolicyUpdateCmd)
	tokenPolicyCmd.AddCommand(tokenPolicyDeleteCmd)
	rootCmd.AddCommand(tokenPolicyCmd)
}
