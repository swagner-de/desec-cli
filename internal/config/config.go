package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Token  string `yaml:"token"`
	Output string `yaml:"output"`
}

// Dir returns the per-user configuration directory.
func Dir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config directory: %w", err)
	}
	return filepath.Join(dir, "desec-cli"), nil
}

// DefaultPath returns the default configuration file path.
func DefaultPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Write writes cfg atomically and ensures the file is only readable by the
// current user.
func Write(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return fmt.Errorf("restricting config directory permissions: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary config file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("restricting config file permissions: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("writing config file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("syncing config file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing config file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		if runtime.GOOS != "windows" {
			return fmt.Errorf("writing config file: %w", err)
		}
		// Windows cannot atomically replace an existing file.
		if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
			return fmt.Errorf("replacing config file: %w", err)
		}
		if err := os.Rename(tmpName, path); err != nil {
			return fmt.Errorf("writing config file: %w", err)
		}
	}

	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("restricting config file permissions: %w", err)
	}

	return nil
}
