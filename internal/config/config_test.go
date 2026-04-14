package config

import (
	"os"
	"path/filepath"
	"testing"
)

func withTempHome(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	orig, _ := os.UserHomeDir()
	t.Setenv("HOME", dir)
	return func() { t.Setenv("HOME", orig) }
}

func TestLoad_MissingFile(t *testing.T) {
	defer withTempHome(t)()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.ProjectName != "" {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	defer withTempHome(t)()

	home, _ := os.UserHomeDir()
	path := filepath.Join(home, configFileName)
	if err := os.WriteFile(path, []byte("not json {{"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ProjectName != "" {
		t.Errorf("expected empty config on bad JSON, got %+v", cfg)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	defer withTempHome(t)()

	want := &Config{ProjectName: "my-project"}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.ProjectName != want.ProjectName {
		t.Errorf("ProjectName: got %q, want %q", got.ProjectName, want.ProjectName)
	}
}

func TestSave_FilePermissions(t *testing.T) {
	defer withTempHome(t)()

	cfg := &Config{ProjectName: "test"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	home, _ := os.UserHomeDir()
	path := filepath.Join(home, configFileName)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file perm: got %o, want 0600", perm)
	}
}
