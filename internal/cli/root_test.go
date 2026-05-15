package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"zjsh/internal/platform"
	"zjsh/internal/provider/config"
	"zjsh/internal/provider/zellij"
	"zjsh/internal/provider/zoxide"
	"zjsh/internal/service"
)

type fakeRunner struct {
	outputs map[string]string
	paths   map[string]string
	errs    map[string]error
	calls   []string
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	if err, ok := f.errs["lookpath:"+file]; ok {
		return "", err
	}
	if path, ok := f.paths[file]; ok {
		return path, nil
	}
	return "/usr/bin/" + file, nil
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := name
	for _, arg := range args {
		key += " " + arg
	}
	f.calls = append(f.calls, key)
	return []byte(f.outputs[key]), nil
}

func (f *fakeRunner) RunInteractive(_ context.Context, name string, args ...string) error {
	key := name
	for _, arg := range args {
		key += " " + arg
	}
	f.calls = append(f.calls, key)
	return nil
}

type fakeEnv struct{ inZellij bool }

func (f fakeEnv) LookupEnv(string) (string, bool) { return "", false }
func (f fakeEnv) InZellij() bool                  { return f.inZellij }

func TestListJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := &fakeRunner{outputs: map[string]string{
		"zellij list-sessions -n": "api [LIVE]\nold [RESURRECT]\n",
		"zoxide query -l":         "/tmp/api\n",
	}}
	app := App{
		Stdout: &stdout,
		Stderr: &stderr,
		Aggregator: service.Aggregator{
			Config: config.Loader{Path: "../../testdata/config/basic.kdl"},
			Zellij: zellij.Provider{Runner: runner},
			Zoxide: zoxide.Provider{Runner: runner},
		},
	}
	if code := app.Run(context.Background(), []string{"list", "--json"}); code != 0 {
		t.Fatalf("Run() code = %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "old") {
		t.Fatalf("expected resurrectable session to be included by default: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "\"name\": \"api\"") {
		t.Fatalf("expected api entry in json: %s", stdout.String())
	}
}

func TestListTextUsesSingleSelectorColumn(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := &fakeRunner{outputs: map[string]string{
		"zellij list-sessions -n": "api [LIVE]\n",
		"zoxide query -l":         "/tmp/api\n/tmp/notes\n",
	}}
	app := App{
		Stdout: &stdout,
		Stderr: &stderr,
		Aggregator: service.Aggregator{
			Config: config.Loader{Path: "../../testdata/config/basic.kdl"},
			Zellij: zellij.Provider{Runner: runner},
			Zoxide: zoxide.Provider{Runner: runner},
		},
	}
	if code := app.Run(context.Background(), []string{"list"}); code != 0 {
		t.Fatalf("Run() code = %d stderr=%s", code, stderr.String())
	}
	got := strings.TrimSpace(stdout.String())
	if strings.Contains(got, "TYPE") || strings.Contains(got, "SOURCE") {
		t.Fatalf("expected selector-only output, got: %s", got)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 selectors, got %d: %q", len(lines), got)
	}
	if lines[0] != "◆ api" {
		t.Fatalf("expected first selector to be project/session name, got %q", lines[0])
	}
	if lines[1] != "→ /tmp/notes" {
		t.Fatalf("expected path selector for zoxide-only entry, got %q", lines[1])
	}
}

func TestConnectResurrectableAttachesByDefault(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := &fakeRunner{outputs: map[string]string{
		"zellij list-sessions -n": "old [RESURRECT]\n",
		"zoxide query -l":         "",
		"zellij attach old":       "",
	}}
	app := App{
		Stdout: &stdout,
		Stderr: &stderr,
		Aggregator: service.Aggregator{
			Config: config.Loader{Path: filepath.Join(t.TempDir(), "missing.kdl")},
			Zellij: zellij.Provider{Runner: runner},
			Zoxide: zoxide.Provider{Runner: runner},
		},
		Launcher: service.Launcher{Runner: runner, Env: fakeEnv{}},
	}
	if code := app.Run(context.Background(), []string{"connect", "old"}); code != 0 {
		t.Fatalf("expected connect to attach resurrectable target, stderr=%q", stderr.String())
	}
	if runner.calls[len(runner.calls)-1] != "zellij attach old" {
		t.Fatalf("unexpected connect command: %#v", runner.calls)
	}
}

func TestConnectRestartsResurrectableByConfiguredPolicy(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := &fakeRunner{outputs: map[string]string{
		"zellij list-sessions -n":   "api [RESURRECT]\n",
		"zoxide query -l":           "",
		"zellij delete-session api": "",
	}}
	app := App{
		Stdout: &stdout,
		Stderr: &stderr,
		Aggregator: service.Aggregator{
			Config: config.Loader{Path: "../../testdata/config/basic.kdl"},
			Zellij: zellij.Provider{Runner: runner},
			Zoxide: zoxide.Provider{Runner: runner},
		},
		Launcher: service.Launcher{Runner: runner, Env: fakeEnv{}},
	}
	if code := app.Run(context.Background(), []string{"connect", "api"}); code != 0 {
		t.Fatalf("expected connect to restart resurrectable session, stderr=%q", stderr.String())
	}
	if len(runner.calls) < 4 {
		t.Fatalf("expected delete and recreate calls, got %#v", runner.calls)
	}
	if runner.calls[len(runner.calls)-2] != "zellij delete-session api" {
		t.Fatalf("expected delete-session before recreate, got %#v", runner.calls)
	}
}

func TestConnectAcceptsSelectorLabel(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := &fakeRunner{outputs: map[string]string{
		"zellij list-sessions -n": "api [LIVE]\n",
		"zoxide query -l":         "",
	}}
	app := App{
		Stdout: &stdout,
		Stderr: &stderr,
		Aggregator: service.Aggregator{
			Config: config.Loader{Path: filepath.Join(t.TempDir(), "missing.kdl")},
			Zellij: zellij.Provider{Runner: runner},
			Zoxide: zoxide.Provider{Runner: runner},
		},
		Launcher: service.Launcher{Runner: runner, Env: fakeEnv{}},
	}
	if code := app.Run(context.Background(), []string{"connect", "● api"}); code != 0 {
		t.Fatalf("expected connect with selector label to succeed, stderr=%q", stderr.String())
	}
	if runner.calls[len(runner.calls)-1] != "zellij attach api" {
		t.Fatalf("unexpected connect command: %#v", runner.calls)
	}
}

func TestDoctorWarnsOnMissingConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := &fakeRunner{}
	app := App{
		Stdout: &stdout,
		Stderr: &stderr,
		Runner: runner,
		Aggregator: service.Aggregator{
			Config: config.Loader{Path: filepath.Join(t.TempDir(), "missing.kdl")},
			Zellij: zellij.Provider{Runner: runner},
			Zoxide: zoxide.Provider{Runner: runner},
		},
	}
	if code := app.Run(context.Background(), []string{"doctor"}); code != 0 {
		t.Fatalf("expected doctor to succeed with warning, stderr=%q stdout=%q", stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "[WARN] config") {
		t.Fatalf("expected missing config warning, got %q", stdout.String())
	}
}

func TestDoctorFailsOnMissingProjectPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "missing" { path "/tmp/definitely-missing-zjsh-path" }`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	runner := &fakeRunner{}
	app := App{
		Stdout: &stdout,
		Stderr: &stderr,
		Runner: runner,
		Aggregator: service.Aggregator{
			Config: config.Loader{Path: configPath},
			Zellij: zellij.Provider{Runner: runner},
			Zoxide: zoxide.Provider{Runner: runner},
		},
	}
	if code := app.Run(context.Background(), []string{"doctor"}); code == 0 {
		t.Fatalf("expected doctor to fail, stdout=%q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "[FAIL] project path") {
		t.Fatalf("expected failing project path check, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "doctor found 1 failing checks") {
		t.Fatalf("expected summary error in stderr, got %q", stderr.String())
	}
}

func TestConfigInitWritesSampleConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := &fakeRunner{}
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	app := App{
		Stdout:     &stdout,
		Stderr:     &stderr,
		Runner:     runner,
		Aggregator: service.Aggregator{Config: config.Loader{}},
	}
	if code := app.Run(context.Background(), []string{"config", "init", "--path", configPath}); code != 0 {
		t.Fatalf("expected config init success, stderr=%q", stderr.String())
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), `project "api"`) {
		t.Fatalf("expected sample config content, got %q", string(data))
	}
	if !strings.Contains(stdout.String(), configPath) {
		t.Fatalf("expected written path in stdout, got %q", stdout.String())
	}
}

func TestConfigInitRefusesOverwrite(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := &fakeRunner{}
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte("defaults {}\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	app := App{
		Stdout:     &stdout,
		Stderr:     &stderr,
		Runner:     runner,
		Aggregator: service.Aggregator{Config: config.Loader{}},
	}
	if code := app.Run(context.Background(), []string{"config", "init", "--path", configPath}); code == 0 {
		t.Fatalf("expected config init overwrite refusal")
	}
	if !strings.Contains(stderr.String(), fmt.Sprintf("config already exists: %s", configPath)) {
		t.Fatalf("expected overwrite refusal, got %q", stderr.String())
	}
}

func TestDoctorFailsWhenBinaryMissing(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := &fakeRunner{errs: map[string]error{"lookpath:zellij": fmt.Errorf("%w: zellij", platform.ErrCommandNotFound)}}
	app := App{
		Stdout: &stdout,
		Stderr: &stderr,
		Runner: runner,
		Aggregator: service.Aggregator{
			Config: config.Loader{Path: filepath.Join(t.TempDir(), "missing.kdl")},
			Zellij: zellij.Provider{Runner: runner},
			Zoxide: zoxide.Provider{Runner: runner},
		},
	}
	if code := app.Run(context.Background(), []string{"doctor"}); code == 0 {
		t.Fatalf("expected doctor failure when zellij is missing")
	}
	if !strings.Contains(stdout.String(), "[FAIL] zellij") {
		t.Fatalf("expected failing zellij check, got %q", stdout.String())
	}
}
