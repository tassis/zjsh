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
