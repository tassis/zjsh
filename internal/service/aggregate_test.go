package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/saweima12/zjsh/internal/domain"
	"github.com/saweima12/zjsh/internal/provider/config"
	"github.com/saweima12/zjsh/internal/provider/zellij"
	"github.com/saweima12/zjsh/internal/provider/zoxide"
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
	if len(entries) != 4 {
		t.Fatalf("expected project and zoxide action entries, got %d", len(entries))
	}
	if entries[0].Type != domain.EntrySession || entries[0].Name != "api" || len(entries[0].Sources) != 2 {
		t.Fatalf("unexpected merged live session entry: %+v", entries[0])
	}
	if !entries[0].RestartOnResurrection {
		t.Fatalf("expected merged restart policy, got %+v", entries[0])
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
	if len(entries) != 2 || entries[0].SessionState != "resurrectable" || !entries[1].CurrentDir {
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
	if len(entries) != 2 || !entries[0].RestartOnResurrection || !entries[1].CurrentDir {
		t.Fatalf("expected inherited restart policy, got %+v", entries)
	}
}

func TestListEntriesDefaultsProjectSessionNameToProjectName(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "project-zjsh" {
  path "/tmp/repos/zjsh"
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	agg := Aggregator{
		Config: config.Loader{Path: configPath},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "project-zjsh [LIVE]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected project and session to merge, got %#v", entries)
	}
	if entries[0].Type != domain.EntrySession || entries[0].SessionName != "project-zjsh" || entries[0].SessionState != domain.SessionStateLive {
		t.Fatalf("expected live session to be primary, got %+v", entries[0])
	}
}

func TestListEntriesUsesExplicitProjectSessionName(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "project-zjsh" {
  path "/tmp/repos/zjsh"
  session "zjsh"
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	agg := Aggregator{
		Config: config.Loader{Path: configPath},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "zjsh [LIVE]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected project and explicit session to merge, got %#v", entries)
	}
	if entries[0].Type != domain.EntrySession || entries[0].Name != "zjsh" || entries[0].SessionName != "zjsh" || entries[0].SessionState != domain.SessionStateLive {
		t.Fatalf("expected explicit live session to be primary, got %+v", entries[0])
	}
}

func TestListEntriesDoesNotMergeExplicitSessionProjectWithProjectNameSession(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "api" {
  path "/tmp/api"
  session "work"
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	agg := Aggregator{
		Config: config.Loader{Path: configPath},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "api [LIVE]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected live session, project, and current dir, got %#v", entries)
	}
	if entries[0].Type != domain.EntrySession || entries[0].SessionName != "api" || entries[0].SessionState != domain.SessionStateLive {
		t.Fatalf("expected api live session to stay separate, got %+v", entries[0])
	}
	if entries[1].Type != domain.EntryProject || entries[1].Name != "api" || entries[1].SessionName != "work" || entries[1].SessionState != domain.SessionStateNone {
		t.Fatalf("expected explicit-session project without live state, got %+v", entries[1])
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
	if len(entries) != 4 {
		t.Fatalf("expected 3 path entries, got %d", len(entries))
	}
	if !entries[0].CurrentDir || entries[1].Path != "/tmp/third" || entries[2].Path != "/tmp/first" || entries[3].Path != "/tmp/second" {
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
	if len(entries) != 4 {
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

func TestListEntriesShowsSessionAndSameBasenameZoxidePath(t *testing.T) {
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
	if len(entries) != 3 {
		t.Fatalf("expected session and zoxide path action entries, got %#v", entries)
	}
	if entries[0].Type != domain.EntrySession || entries[0].SessionState != domain.SessionStateLive {
		t.Fatalf("expected live session first, got %+v", entries[0])
	}
	if !entries[1].CurrentDir || entries[2].Type != domain.EntryPath || entries[2].Path != "/tmp/foo" || entries[2].SessionName != "" {
		t.Fatalf("expected current dir then separate zoxide path action, got %+v / %+v", entries[1], entries[2])
	}
}

func TestListEntriesShowsProjectSessionAndSameBasenameZoxidePath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "project-foo" {
  path "/tmp/project/foo"
  session "foo"
  startup "nvim ."
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	agg := Aggregator{
		Config: config.Loader{Path: configPath},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "foo [LIVE]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "/tmp/other/foo\n",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected project and zoxide path action entries, got %#v", entries)
	}
	entry := entries[0]
	if entry.Type != domain.EntrySession || entry.Name != "foo" || entry.Path != "/tmp/project/foo" {
		t.Fatalf("expected live session to win merged identity, got %+v", entry)
	}
	if entry.SessionState != domain.SessionStateLive || entry.SessionName != "foo" {
		t.Fatalf("expected live session state to be preserved, got %+v", entry)
	}
	if len(entry.Sources) != 2 {
		t.Fatalf("expected merged sources, got %+v", entry)
	}
	if !entries[1].CurrentDir || entries[2].Type != domain.EntryPath || entries[2].Path != "/tmp/other/foo" || entries[2].SessionName != "" {
		t.Fatalf("expected current dir then separate zoxide path action, got %+v / %+v", entries[1], entries[2])
	}
}

func TestListEntriesKeepsProjectsWithSameSessionBasenameSeparate(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "config-zjsh" {
  path "/tmp/config/zjsh"
}

project "project-zjsh" {
  path "/tmp/repos/zjsh"
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
			"zoxide query -l": "/tmp/repos/zjsh\n",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected both projects to stay separate, got %#v", entries)
	}
	if entries[0].Name != "config-zjsh" || entries[0].Path != "/tmp/config/zjsh" {
		t.Fatalf("unexpected first project: %+v", entries[0])
	}
	if entries[1].Name != "project-zjsh" || entries[1].Path != "/tmp/repos/zjsh" {
		t.Fatalf("unexpected second project: %+v", entries[1])
	}
	if !entries[2].CurrentDir || entries[3].Type != domain.EntryPath || entries[3].Path != "/tmp/repos/zjsh" {
		t.Fatalf("expected current dir then separate zoxide path action, got %+v / %+v", entries[2], entries[3])
	}
}

func TestListEntriesShowsProjectSessionAndSamePathZoxideAction(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "project-zjsh" {
  path "/tmp/repos/zjsh"
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	agg := Aggregator{
		Config: config.Loader{Path: configPath},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "zjsh [LIVE]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "/tmp/repos/zjsh\n",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected project, live session, and zoxide path action, got %#v", entries)
	}
	if entries[0].Type != domain.EntrySession || entries[0].SessionName != "zjsh" || entries[0].SessionState != domain.SessionStateLive {
		t.Fatalf("expected live session first, got %+v", entries[0])
	}
	if entries[1].Type != domain.EntryProject || entries[1].Name != "project-zjsh" || entries[1].SessionName != "project-zjsh" {
		t.Fatalf("expected project to win path identity, got %+v", entries[0])
	}
	if !entries[2].CurrentDir || entries[3].Type != domain.EntryPath || entries[3].Path != "/tmp/repos/zjsh" {
		t.Fatalf("expected current dir then separate zoxide path action, got %+v / %+v", entries[2], entries[3])
	}
}

func TestListEntriesPrefersProjectWhenSessionMerges(t *testing.T) {
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
	if len(entries) != 3 {
		t.Fatalf("expected project plus zoxide path action, got %#v", entries)
	}
	entry := entries[0]
	if entry.Type != domain.EntrySession || entry.Name != "api-session" || entry.Startup != "nvim ." {
		t.Fatalf("expected live session to win merged identity while preserving project config, got %+v", entry)
	}
	if entry.SessionState != domain.SessionStateLive {
		t.Fatalf("expected zellij session state to be preserved, got %+v", entry)
	}
	if len(entry.Sources) != 2 {
		t.Fatalf("expected merged sources, got %+v", entry)
	}
	if !entries[1].CurrentDir || entries[2].Type != domain.EntryPath || entries[2].Path != "/tmp/api" {
		t.Fatalf("expected current dir then separate zoxide path action, got %+v / %+v", entries[1], entries[2])
	}
}

func TestListEntriesKeepsCurrentDirectorySeparateFromZoxidePath(t *testing.T) {
	agg := Aggregator{
		CWD:    "/tmp/here",
		Config: config.Loader{Path: filepath.Join(t.TempDir(), "missing.kdl")},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "/tmp/here\n/tmp/other\n",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected current dir plus two path entries, got %#v", entries)
	}
	if !entries[0].CurrentDir || entries[0].Path != "/tmp/here" {
		t.Fatalf("expected current dir before zoxide paths, got %+v", entries[0])
	}
	if entries[1].CurrentDir || entries[1].Path != "/tmp/here" {
		t.Fatalf("expected separate matching zoxide path, got %+v", entries[1])
	}
	if entries[2].CurrentDir || entries[2].Path != "/tmp/other" {
		t.Fatalf("expected ordinary zoxide path last, got %+v", entries[2])
	}
	if len(entries[0].Sources) != 1 || entries[0].Sources[0] != "cwd" {
		t.Fatalf("expected current dir to only include cwd source, got %+v", entries[0].Sources)
	}
}

func TestListEntriesSupportsCWDProjectTemplate(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.kdl")
	err := os.WriteFile(configPath, []byte(`project "here" {
  cwd true
  session "work"
  layout "compact"
}
`), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	agg := Aggregator{
		CWD:    "/tmp/runtime",
		Config: config.Loader{Path: configPath},
		Zellij: zellij.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zellij list-sessions -n": "work [LIVE]\n",
		}}},
		Zoxide: zoxide.Provider{Runner: fakeRunner{outputs: map[string]string{
			"zoxide query -l": "",
		}}},
	}
	entries, err := agg.ListEntries(context.Background(), true)
	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected cwd project and current dir, got %#v", entries)
	}
	entry := entries[0]
	if entry.Type != domain.EntrySession || entry.Name != "work" || entry.Path != "/tmp/runtime" || entry.SessionName != "work" {
		t.Fatalf("unexpected cwd project entry: %+v", entry)
	}
	if entry.Layout != "compact" || entry.SessionState != domain.SessionStateLive || entry.SortRank != liveSessionSortRank {
		t.Fatalf("expected live cwd project with layout, got %+v", entry)
	}
}
