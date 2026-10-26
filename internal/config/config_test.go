package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteEnforcesPrivatePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose POSIX file modes")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("token: old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}

	if err := Write(path, &Config{Token: "secret", Output: "table"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("config mode = %o, want 0600", got)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "token: secret\noutput: table\n" {
		t.Fatalf("unexpected config contents: %q", data)
	}
}

func TestWriteCreatesPrivateFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose POSIX file modes")
	}

	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	if err := Write(path, &Config{Token: "secret"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("config mode = %o, want 0600", got)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0700 {
		t.Fatalf("config directory mode = %o, want 0700", got)
	}
}

func TestWritePreservesExistingConfigDirPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose POSIX file modes")
	}

	dir := filepath.Join(t.TempDir(), "shared")
	if err := os.Mkdir(dir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0750); err != nil {
		t.Fatal(err)
	}

	if err := Write(filepath.Join(dir, "config.yaml"), &Config{Token: "secret"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0750 {
		t.Fatalf("config directory mode = %o, want 0750", got)
	}
}
