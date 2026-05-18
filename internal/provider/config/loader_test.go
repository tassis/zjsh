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
  icon_project "P"
  icon_session "S"
  icon_resurrectable "R"
  icon_path "Z"
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
	if config.Defaults.Icons.Project != "P" || config.Defaults.Icons.Session != "S" || config.Defaults.Icons.Resurrectable != "R" || config.Defaults.Icons.Path != "Z" {
		t.Fatalf("unexpected configured icons: %+v", config.Defaults.Icons)
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

func TestParseConfigIgnoresUnknownFields(t *testing.T) {
	for _, field := range []string{"attach", "session_name"} {
		_, err := parseConfig("defaults {\n  " + field + " true\n}\n")
		if err != nil {
			t.Fatalf("expected unknown key %q to be ignored, got %v", field, err)
		}
	}
}

func TestParseConfigSupportsCWDProject(t *testing.T) {
	config, err := parseConfig(`project "here" {
  cwd true
  session "work"
  layout "compact"
}
`)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	project := config.Projects[0]
	if !project.CWD || project.Path != "" || project.Session != "work" || project.Layout != "compact" {
		t.Fatalf("unexpected cwd project: %+v", project)
	}
}

func TestParseConfigRejectsInvalidProjectPathModes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "missing path and cwd",
			input: `project "here" {
}
`,
		},
		{
			name: "path and cwd",
			input: `project "here" {
  path "/tmp/here"
  cwd true
}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseConfig(tt.input)
			if err == nil {
				t.Fatalf("expected parse error")
			}
		})
	}
}

func TestParseConfigIgnoresUnknownTopLevelNodes(t *testing.T) {
	_, err := parseConfig(`plugin "unused" {
  enabled true
}

project "api" {
  path "/tmp/api"
}
`)
	if err != nil {
		t.Fatalf("expected unknown top-level node to be ignored, got %v", err)
	}
}

func TestParseConfigIgnoresUnknownProjectFields(t *testing.T) {
	config, err := parseConfig(`project "api" {
  path "/tmp/api"
  removed_field "ignored"
}
`)
	if err != nil {
		t.Fatalf("expected unknown project field to be ignored, got %v", err)
	}
	if len(config.Projects) != 1 || config.Projects[0].Name != "api" {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestParseConfigRejectsInvalidKDL(t *testing.T) {
	_, err := parseConfig(`project "api" {
  path "/tmp/api"
`)
	if err == nil {
		t.Fatalf("expected invalid KDL syntax error")
	}
}

func TestParseConfigRejectsInvalidBooleanValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "invalid cwd",
			input: `project "here" {
  cwd "yes"
}
`,
		},
		{
			name: "invalid project restart_on_resurrection",
			input: `project "api" {
  path "/tmp/api"
  restart_on_resurrection "yes"
}
`,
		},
		{
			name: "invalid defaults restart_on_resurrection",
			input: `defaults {
  restart_on_resurrection "yes"
}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseConfig(tt.input)
			if err == nil {
				t.Fatalf("expected invalid boolean error")
			}
		})
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
	if config.Defaults.Shell != "sh" || config.Defaults.Icons.Project != "◆" {
		t.Fatalf("expected default config values, got %+v", config.Defaults)
	}
}
