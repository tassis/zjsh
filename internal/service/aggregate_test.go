package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"zjsh/internal/domain"
	"zjsh/internal/provider/config"
	"zjsh/internal/provider/zellij"
	"zjsh/internal/provider/zoxide"
)

type fakeRunner struct {
	outputs map[string]string
}

func (f fakeRunner) LookPath(file string) (string, error) {
	return "/usr/bin/" + file, nil
}

func (f fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := name
	for _, arg := range args {
		key += " " + arg
	}
	return []byte(f.outputs[key]), nil
}

func (f fakeRunner) RunInteractive(context.Context, string, ...string) error {
	return nil
}

func TestListEntriesCanExcludeResurrectable(t *testing.T) {
	agg := Aggregator{
		Config: config.Loader{Path: "../../testdata/config/basic.kdl"},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "api [LIVE]\napi-old [RESURRECT]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "/tmp/api\n/tmp/notes\n",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), false)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	for _, entry := range entries {
		if entry.SessionState == "resurrectable" {
			t.Fatalf("unexpected resurrectable entry in default list: %+v", entry)
		}
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after dedupe, got %d", len(entries))
	}
	if entries[0].Name != "api" || len(entries[0].Sources) != 3 {
		t.Fatalf("unexpected merged project entry: %+v", entries[0])
	}
	if !entries[0].RestartOnResurrection {
		t.Fatalf("expected merged project restart policy, got %+v", entries[0])
	}
}

func TestListEntriesIncludeResurrectable(t *testing.T) {
	agg := Aggregator{
		Config: config.Loader{Path: filepath.Join(t.TempDir(), "missing.kdl")},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "api-old [RESURRECT]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 1 || entries[0].SessionState != "resurrectable" {
		t.Fatalf("unexpected resurrectable list: %+v", entries)
	}
}

func TestListEntriesUsesDefaultRestartPolicy(t *testing.T) {
	agg := Aggregator{
		Config: config.Loader{Path: "../../testdata/config/default-restart.kdl"},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "api [RESURRECT]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 1 || !entries[0].RestartOnResurrection {
		t.Fatalf("expected inherited restart policy, got %+v", entries)
	}
}

func TestListEntriesPreservesZoxideOrderWithinPathEntries(t *testing.T) {
	agg := Aggregator{
		Config: config.Loader{Path: filepath.Join(t.TempDir(), "missing.kdl")},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "/tmp/third\n/tmp/first\n/tmp/second\n",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), false)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 path entries, got %d", len(entries))
	}
	if entries[0].Path != "/tmp/third" || entries[1].Path != "/tmp/first" || entries[2].Path != "/tmp/second" {
		t.Fatalf("expected zoxide order to be preserved, got %#v", entries)
	}
}

func TestListEntriesKeepsSameBasenamePathsSeparate(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "foo" {
  path "/tmp/work/foo"
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	agg := Aggregator{
		Config: config.Loader{Path: configPath},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "/tmp/other/foo\n/tmp/another/foo\n",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected project and both zoxide paths to stay separate, got %#v", entries)
	}
	if entries[0].Type != domain.EntryProject || entries[0].Path != "/tmp/work/foo" {
		t.Fatalf("expected project entry first, got %+v", entries[0])
	}
	for _, entry := range entries[1:] {
		if entry.Type != domain.EntryPath || entry.SessionName != "" {
			t.Fatalf("expected path-only entry without session alias, got %+v", entry)
		}
	}
}

func TestListEntriesPrefersLiveSessionOverSameBasenamePath(t *testing.T) {
	agg := Aggregator{
		Config: config.Loader{Path: filepath.Join(t.TempDir(), "missing.kdl")},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "foo [LIVE]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "/tmp/foo\n",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected session and zoxide path to merge, got %#v", entries)
	}
	entry := entries[0]
	if entry.Type != domain.EntrySession || entry.SessionState != domain.SessionStateLive {
		t.Fatalf("expected live session to win, got %+v", entry)
	}
	if entry.Path != "/tmp/foo" || entry.SessionName != "foo" {
		t.Fatalf("expected zoxide path to be preserved on session entry, got %+v", entry)
	}
	if len(entry.Sources) != 2 {
		t.Fatalf("expected merged sources, got %+v", entry)
	}
}

func TestListEntriesPrefersProjectWhenSourcesMerge(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "api" {
  path "/tmp/api"
  session "api-session"
  startup "nvim ."
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	agg := Aggregator{
		Config: config.Loader{Path: configPath},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "api-session [LIVE]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "/tmp/api\n",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected merged entry, got %#v", entries)
	}
	entry := entries[0]
	if entry.Type != domain.EntryProject || entry.Name != "api" || entry.Startup != "nvim ." {
		t.Fatalf("expected project fields to win, got %+v", entry)
	}
	if entry.SessionState != domain.SessionStateLive {
		t.Fatalf("expected zellij session state to be preserved, got %+v", entry)
	}
	if len(entry.Sources) != 3 {
		t.Fatalf("expected merged sources, got %+v", entry)
	}
}
