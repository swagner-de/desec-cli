package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/swagner-de/desec-cli/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage desec-cli configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter your deSEC API token: ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)
		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		path := config.DefaultPath()
		cfg := &config.Config{
			Token:  token,
			Output: "table",
		}

		if err := config.Write(path, cfg); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Configuration written to %s\n", path)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)
}
