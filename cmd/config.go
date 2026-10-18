package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/swagner-de/desec-cli/internal/config"
	"golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage desec-cli configuration",
}

var configInitForce bool

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readToken(cmd)
		if err != nil {
			return err
		}
		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		path, err := config.DefaultPath()
		if err != nil {
			return err
		}
		if err := ensureConfigWritable(path, configInitForce); err != nil {
			return err
		}
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
	configInitCmd.Flags().BoolVar(&configInitForce, "force", false, "Overwrite an existing configuration file")
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)
}

func readToken(cmd *cobra.Command) (string, error) {
	if f, ok := cmd.InOrStdin().(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		fmt.Fprint(cmd.ErrOrStderr(), "Enter your deSEC API token: ")
		raw, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", fmt.Errorf("reading token: %w", err)
		}
		return strings.TrimSpace(string(raw)), nil
	}

	reader := bufio.NewReader(cmd.InOrStdin())
	token, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("reading token: %w", err)
	}
	return strings.TrimSpace(token), nil
}

func ensureConfigWritable(path string, force bool) error {
	if force {
		return nil
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("configuration already exists at %s; use --force to overwrite", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking configuration file: %w", err)
	}
	return nil
}
