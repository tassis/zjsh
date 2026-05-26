package zellij

import (
	"context"
	"fmt"
	"testing"

	"github.com/tassis/zjsh/internal/domain"
)

func TestParseSessions(t *testing.T) {
	input := `
api [LIVE]
blog [EXITED]
dev [RESURRECT]
`
	sessions := ParseSessions(input)
	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(sessions))
	}
	if sessions[0].Name != "api" || sessions[0].State != domain.SessionStateLive {
		t.Fatalf("unexpected live session: %+v", sessions[0])
	}
	if sessions[1].State != domain.SessionStateResurrectable {
		t.Fatalf("unexpected exited state: %+v", sessions[1])
	}
	if sessions[2].State != domain.SessionStateResurrectable {
		t.Fatalf("unexpected resurrect state: %+v", sessions[2])
	}
}

func TestParseSessionsStripsANSI(t *testing.T) {
	input := "\x1b[32;1mcubic-weasel\x1b[m [Created \x1b[35;1m1day\x1b[m ago]\n"
	sessions := ParseSessions(input)
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Name != "cubic-weasel" {
		t.Fatalf("expected stripped session name, got %q", sessions[0].Name)
	}
}

type fakeRunner struct {
	out []byte
	err error
}

func (f fakeRunner) LookPath(string) (string, error) {
	return "/usr/bin/zellij", nil
}

func (f fakeRunner) Run(context.Context, string, ...string) ([]byte, error) {
	return f.out, f.err
}

func (f fakeRunner) RunInteractive(context.Context, string, ...string) error {
	return nil
}

func TestListSessionsReturnsEmptyWhenNoActiveSessions(t *testing.T) {
	provider := Provider{Runner: fakeRunner{
		out: []byte("No active zellij sessions found.\n"),
		err: fmt.Errorf("exit status 1"),
	}}
	sessions, err := provider.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected empty sessions, got %+v", sessions)
	}
}
