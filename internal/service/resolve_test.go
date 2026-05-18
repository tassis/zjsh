package service

import (
	"testing"

	"github.com/saweima12/zjsh/internal/domain"
)

func TestResolveExactOrder(t *testing.T) {
	entries := []domain.Entry{
		{Name: "api", Type: domain.EntrySession, SessionName: "api", SessionState: domain.SessionStateLive, Score: 300},
		{Name: "api", Type: domain.EntryProject, SessionName: "workspace", Path: "/tmp/api", Score: 400},
	}
	entry, err := ResolveExact(entries, "api")
	if err != nil {
		t.Fatalf("ResolveExact() error = %v", err)
	}
	if entry.Type != domain.EntryProject {
		t.Fatalf("expected project match first, got %+v", entry)
	}
}

func TestResolveExactCurrentDirBeatsDotNamedProject(t *testing.T) {
	entries := []domain.Entry{
		{Name: ".", Type: domain.EntryProject, SessionName: "dot-project", Path: "/tmp/project"},
		{Name: ".", Type: domain.EntryPath, Path: "/tmp/here", CurrentDir: true},
	}
	entry, err := ResolveExact(entries, ".")
	if err != nil {
		t.Fatalf("ResolveExact() error = %v", err)
	}
	if !entry.CurrentDir || entry.Path != "/tmp/here" {
		t.Fatalf("expected current dir entry, got %+v", entry)
	}
}

func TestResolveExactAllowsResurrectableByDefault(t *testing.T) {
	entries := []domain.Entry{{Name: "old", Type: domain.EntrySession, SessionName: "old", SessionState: domain.SessionStateResurrectable}}
	entry, err := ResolveExact(entries, "old")
	if err != nil {
		t.Fatalf("ResolveExact() error = %v", err)
	}
	if entry.SessionState != domain.SessionStateResurrectable {
		t.Fatalf("expected resurrectable entry, got %+v", entry)
	}
}

func TestResolveExactAllowsRestartOnResurrection(t *testing.T) {
	entries := []domain.Entry{{
		Name:                  "api",
		Type:                  domain.EntryProject,
		SessionName:           "api",
		SessionState:          domain.SessionStateResurrectable,
		RestartOnResurrection: true,
	}}
	entry, err := ResolveExact(entries, "api")
	if err != nil {
		t.Fatalf("ResolveExact() error = %v", err)
	}
	if !entry.RestartOnResurrection {
		t.Fatalf("expected restart_on_resurrection entry, got %+v", entry)
	}
}

func TestResolveExactPrefersZoxidePathActionForFullPath(t *testing.T) {
	entries := []domain.Entry{
		{Name: "api", Type: domain.EntryProject, SessionName: "api", Path: "/tmp/api"},
		{Name: "api", Type: domain.EntryPath, Path: "/tmp/api"},
	}
	entry, err := ResolveExact(entries, "/tmp/api")
	if err != nil {
		t.Fatalf("ResolveExact() error = %v", err)
	}
	if entry.Type != domain.EntryPath {
		t.Fatalf("expected zoxide path action for full path, got %+v", entry)
	}
}

func TestPrepareConnectEntryUsesBasenameForFreeZoxideSession(t *testing.T) {
	entry := PrepareConnectEntry(domain.Entry{Type: domain.EntryPath, Path: "/tmp/foo"}, nil)
	if entry.SessionName != "foo" {
		t.Fatalf("expected basename session name, got %+v", entry)
	}
}

func TestPrepareConnectEntryUsesHashFallbackForReservedZoxideSession(t *testing.T) {
	entry := PrepareConnectEntry(
		domain.Entry{Type: domain.EntryPath, Path: "/tmp/foo"},
		[]domain.Entry{{Type: domain.EntrySession, SessionName: "foo"}},
	)
	expected := "foo-" + shortPathHash("/tmp/foo")
	if entry.SessionName != expected {
		t.Fatalf("expected fallback session name %q, got %+v", expected, entry)
	}
}
