package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfig(t *testing.T) {
	config, err := parseConfig(`
defaults {
  shell "bash"
  restart_on_resurrection true
}

project "api" {
  path "/tmp/api"
  session "api"
  startup "nvim ."
  layout "compact"
}

project "ops" {
  path "/tmp/ops"
  layout_file "/tmp/ops.kdl"
}

project "restartable" {
  path "/tmp/restartable"
  restart_on_resurrection true
}

project "inherits" {
  path "/tmp/inherits"
}
`)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if config.Defaults.Shell != "bash" {
		t.Fatalf("expected shell bash, got %q", config.Defaults.Shell)
	}
	if !config.Defaults.RestartOnResurrection {
		t.Fatalf("expected defaults restart_on_resurrection=true")
	}
	if len(config.Projects) != 4 {
		t.Fatalf("expected 4 projects, got %d", len(config.Projects))
	}
	if config.Projects[0].Name != "api" || config.Projects[0].Layout != "compact" {
		t.Fatalf("unexpected first project: %+v", config.Projects[0])
	}
	if config.Projects[1].LayoutFile != "/tmp/ops.kdl" {
		t.Fatalf("unexpected second project: %+v", config.Projects[1])
	}
	if config.Projects[2].RestartOnResurrection == nil || !*config.Projects[2].RestartOnResurrection {
		t.Fatalf("expected restart_on_resurrection for third project: %+v", config.Projects[2])
	}
	if config.Projects[3].RestartOnResurrection != nil {
		t.Fatalf("expected fourth project to inherit defaults: %+v", config.Projects[3])
	}
}

func TestParseConfigRejectsRemovedDefaultsFields(t *testing.T) {
	for _, field := range []string{"attach", "session_name"} {
		_, err := parseConfig("defaults {\n  " + field + " true\n}\n")
		if err == nil {
			t.Fatalf("expected unsupported key error for %q", field)
		}
	}
}

func TestLoadExpandsProjectHomePaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "api" {
  path "~/work/api"
  layout_file "~/.config/zellij/layouts/api.kdl"
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	config, err := Loader{Path: configPath}.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.Projects[0].Path != filepath.Join(home, "work", "api") {
		t.Fatalf("expected expanded project path, got %q", config.Projects[0].Path)
	}
	if config.Projects[0].LayoutFile != filepath.Join(home, ".config", "zellij", "layouts", "api.kdl") {
		t.Fatalf("expected expanded layout file, got %q", config.Projects[0].LayoutFile)
	}
}

func TestLoadKeepsRelativeProjectPaths(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "api" {
  path "work/api"
  layout_file "layouts/api.kdl"
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	config, err := Loader{Path: configPath}.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.Projects[0].Path != "work/api" {
		t.Fatalf("expected relative project path unchanged, got %q", config.Projects[0].Path)
	}
	if config.Projects[0].LayoutFile != "layouts/api.kdl" {
		t.Fatalf("expected relative layout file unchanged, got %q", config.Projects[0].LayoutFile)
	}
}

func TestLoadMissingConfigReturnsEmpty(t *testing.T) {
	loader := Loader{Path: "/tmp/definitely-missing-zjsh-config.kdl"}
	config, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(config.Projects) != 0 {
		t.Fatalf("expected empty projects, got %d", len(config.Projects))
	}
}
