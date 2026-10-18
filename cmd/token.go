package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/swagner-de/desec-cli/internal/client"
	"github.com/swagner-de/desec-cli/internal/output"
	"golang.org/x/term"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage API tokens",
}

var tokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		tokens, err := c.ListTokens()
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			headers := []string{"ID", "NAME", "VALID", "CREATED", "LAST USED"}
			var rows [][]string
			for _, t := range tokens {
				rows = append(rows, []string{
					t.ID,
					t.Name,
					fmt.Sprintf("%v", t.IsValid),
					t.Created,
					t.LastUsed,
				})
			}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, tokens)
	},
}

var tokenGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get token details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		t, err := c.GetToken(args[0])
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			headers := []string{"ID", "NAME", "VALID", "MANAGE TOKENS", "CREATE DOMAIN", "DELETE DOMAIN", "SUBNETS"}
			rows := [][]string{{
				t.ID,
				t.Name,
				fmt.Sprintf("%v", t.IsValid),
				fmt.Sprintf("%v", t.PermManageTokens),
				fmt.Sprintf("%v", t.PermCreateDomain),
				fmt.Sprintf("%v", t.PermDeleteDomain),
				strings.Join(t.AllowedSubnets, ", "),
			}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, t)
	},
}

var tokenCreateName string
var tokenCreateSubnets []string
var tokenCreatePermManageTokens bool
var tokenCreatePermCreateDomain bool
var tokenCreatePermDeleteDomain bool
var tokenCreateMaxAge string
var tokenCreateMaxUnusedPeriod string
var tokenCreateShowSecret bool

var tokenDurationPattern = regexp.MustCompile(`^(\d+ )?((\d{1,2}:)?\d{1,2}:\d{2}|\d+)(\.\d{1,6})?$`)

var tokenCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateTokenDuration("--max-age", tokenCreateMaxAge); err != nil {
			return err
		}
		if err := validateTokenDuration("--max-unused-period", tokenCreateMaxUnusedPeriod); err != nil {
			return err
		}
		c := client.New(getToken())
		tc := &client.TokenCreate{
			Name:            tokenCreateName,
			AllowedSubnets:  tokenCreateSubnets,
			MaxAge:          tokenCreateMaxAge,
			MaxUnusedPeriod: tokenCreateMaxUnusedPeriod,
		}
		if cmd.Flags().Changed("perm-manage-tokens") {
			v := tokenCreatePermManageTokens
			tc.PermManageTokens = &v
		}
		if cmd.Flags().Changed("perm-create-domain") {
			v := tokenCreatePermCreateDomain
			tc.PermCreateDomain = &v
		}
		if cmd.Flags().Changed("perm-delete-domain") {
			v := tokenCreatePermDeleteDomain
			tc.PermDeleteDomain = &v
		}
		t, err := c.CreateToken(tc)
		if err != nil {
			return err
		}
		if shouldPrintTokenSecret(cmd, tokenCreateShowSecret) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Token created. Secret (shown once): %s\n", t.Token)
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), "Token created. Secret not printed because stderr is not a terminal; use --show-secret or --output json if you need it.")
		}
		if outputFormat == "table" {
			headers := []string{"ID", "NAME", "VALID", "MANAGE TOKENS", "CREATE DOMAIN", "DELETE DOMAIN", "SUBNETS"}
			rows := [][]string{{
				t.ID,
				t.Name,
				fmt.Sprintf("%v", t.IsValid),
				fmt.Sprintf("%v", t.PermManageTokens),
				fmt.Sprintf("%v", t.PermCreateDomain),
				fmt.Sprintf("%v", t.PermDeleteDomain),
				strings.Join(t.AllowedSubnets, ", "),
			}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, t)
	},
}

var tokenUpdateName string
var tokenUpdateSubnets []string
var tokenUpdatePermManageTokens bool
var tokenUpdatePermCreateDomain bool
var tokenUpdatePermDeleteDomain bool

var tokenUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(getToken())
		update := map[string]any{}
		if cmd.Flags().Changed("name") {
			update["name"] = tokenUpdateName
		}
		if cmd.Flags().Changed("subnet") {
			update["allowed_subnets"] = tokenUpdateSubnets
		}
		if cmd.Flags().Changed("perm-manage-tokens") {
			update["perm_manage_tokens"] = tokenUpdatePermManageTokens
		}
		if cmd.Flags().Changed("perm-create-domain") {
			update["perm_create_domain"] = tokenUpdatePermCreateDomain
		}
		if cmd.Flags().Changed("perm-delete-domain") {
			update["perm_delete_domain"] = tokenUpdatePermDeleteDomain
		}
		if len(update) == 0 {
			return fmt.Errorf("nothing to update — specify at least one flag")
		}
		t, err := c.UpdateToken(args[0], update)
		if err != nil {
			return err
		}
		if outputFormat == "table" {
			fmt.Fprintln(os.Stderr, "Token updated.")
			headers := []string{"ID", "NAME", "VALID", "MANAGE TOKENS", "CREATE DOMAIN", "DELETE DOMAIN", "SUBNETS"}
			rows := [][]string{{
				t.ID,
				t.Name,
				fmt.Sprintf("%v", t.IsValid),
				fmt.Sprintf("%v", t.PermManageTokens),
				fmt.Sprintf("%v", t.PermCreateDomain),
				fmt.Sprintf("%v", t.PermDeleteDomain),
				strings.Join(t.AllowedSubnets, ", "),
			}}
			return output.PrintTable(headers, rows)
		}
		return output.Print(outputFormat, t)
	},
}

var tokenDeleteYes bool

var tokenDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tokenDeleteYes {
			fmt.Fprintf(os.Stderr, "Delete token %s? [y/N]: ", args[0])
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(answer)) != "y" {
				fmt.Fprintln(os.Stderr, "Aborted.")
				return nil
			}
		}
		c := client.New(getToken())
		if err := c.DeleteToken(args[0]); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Token %s deleted.\n", args[0])
		return nil
	},
}

func init() {
	tokenCreateCmd.Flags().StringVar(&tokenCreateName, "name", "", "Token name")
	tokenCreateCmd.Flags().StringArrayVar(&tokenCreateSubnets, "subnet", nil, "Allowed subnet (repeatable)")
	tokenCreateCmd.Flags().BoolVar(&tokenCreatePermManageTokens, "perm-manage-tokens", false, "Allow managing tokens")
	tokenCreateCmd.Flags().BoolVar(&tokenCreatePermCreateDomain, "perm-create-domain", false, "Allow creating domains")
	tokenCreateCmd.Flags().BoolVar(&tokenCreatePermDeleteDomain, "perm-delete-domain", false, "Allow deleting domains")
	tokenCreateCmd.Flags().StringVar(&tokenCreateMaxAge, "max-age", "", "Maximum token age (deSEC duration, e.g. \"30 00:00:00\")")
	tokenCreateCmd.Flags().StringVar(&tokenCreateMaxUnusedPeriod, "max-unused-period", "", "Maximum unused period (deSEC duration, e.g. \"7 00:00:00\")")
	tokenCreateCmd.Flags().BoolVar(&tokenCreateShowSecret, "show-secret", false, "Print a newly created token secret even when stderr is not a terminal")

	tokenUpdateCmd.Flags().StringVar(&tokenUpdateName, "name", "", "Token name")
	tokenUpdateCmd.Flags().StringArrayVar(&tokenUpdateSubnets, "subnet", nil, "Allowed subnet (repeatable, replaces all)")
	tokenUpdateCmd.Flags().BoolVar(&tokenUpdatePermManageTokens, "perm-manage-tokens", false, "Allow managing tokens")
	tokenUpdateCmd.Flags().BoolVar(&tokenUpdatePermCreateDomain, "perm-create-domain", false, "Allow creating domains")
	tokenUpdateCmd.Flags().BoolVar(&tokenUpdatePermDeleteDomain, "perm-delete-domain", false, "Allow deleting domains")

	tokenDeleteCmd.Flags().BoolVarP(&tokenDeleteYes, "yes", "y", false, "Skip confirmation prompt")

	tokenCmd.AddCommand(tokenListCmd)
	tokenCmd.AddCommand(tokenGetCmd)
	tokenCmd.AddCommand(tokenCreateCmd)
	tokenCmd.AddCommand(tokenUpdateCmd)
	tokenCmd.AddCommand(tokenDeleteCmd)
	rootCmd.AddCommand(tokenCmd)
}

func shouldPrintTokenSecret(cmd *cobra.Command, force bool) bool {
	if force {
		return true
	}
	f, ok := cmd.ErrOrStderr().(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

func validateTokenDuration(name, value string) error {
	if value == "" {
		return nil
	}
	if !tokenDurationPattern.MatchString(value) {
		return fmt.Errorf("%s must use the deSEC duration format \"[DD] [[HH:]MM:]ss[.uuuuuu]\" (for example \"30 00:00:00\")", name)
	}
	return nil
}
