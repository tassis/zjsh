package service

import (
	"context"
	"strings"
	"testing"

	"github.com/saweima12/zjsh/internal/domain"
)

type recordingRunner struct {
	calls []string
}

func (r *recordingRunner) LookPath(file string) (string, error) {
	return "/usr/bin/" + file, nil
}

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	call := name
	for _, arg := range args {
		call += " " + arg
	}
	r.calls = append(r.calls, call)
	return nil, nil
}

func (r *recordingRunner) RunInteractive(_ context.Context, name string, args ...string) error {
	call := name
	for _, arg := range args {
		call += " " + arg
	}
	r.calls = append(r.calls, call)
	return nil
}

type staticEnv struct{ inZellij bool }

func (e staticEnv) LookupEnv(string) (string, bool) { return "", false }
func (e staticEnv) InZellij() bool                  { return e.inZellij }

func TestLauncherConnectLiveSessionOutsideZellij(t *testing.T) {
	runner := &recordingRunner{}
	launcher := Launcher{Runner: runner, Env: staticEnv{}}
	err := launcher.Connect(context.Background(), domain.Entry{SessionName: "api", SessionState: domain.SessionStateLive})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if runner.calls[0] != "zellij attach api" {
		t.Fatalf("unexpected call: %#v", runner.calls)
	}
}

func TestLauncherConnectLiveSessionInsideZellij(t *testing.T) {
	runner := &recordingRunner{}
	launcher := Launcher{Runner: runner, Env: staticEnv{inZellij: true}}
	err := launcher.Connect(context.Background(), domain.Entry{SessionName: "api", SessionState: domain.SessionStateLive})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if runner.calls[0] != "zellij action switch-session api" {
		t.Fatalf("unexpected call: %#v", runner.calls)
	}
}

func TestLauncherCreateSessionWithStartup(t *testing.T) {
	runner := &recordingRunner{}
	launcher := Launcher{Runner: runner, Env: staticEnv{}}
	err := launcher.Connect(context.Background(), domain.Entry{Type: domain.EntryProject, Name: "api", SessionName: "api", Path: "/tmp/api", Shell: "bash", Startup: "nvim ."})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if len(runner.calls) != 1 || runner.calls[0] == "" {
		t.Fatalf("expected one command, got %#v", runner.calls)
	}
	if !strings.HasPrefix(runner.calls[0], "zellij -s api --layout-string ") {
		t.Fatalf("unexpected startup command: %q", runner.calls[0])
	}
	if !strings.Contains(runner.calls[0], `nvim .`) {
		t.Fatalf("expected startup command in layout string: %q", runner.calls[0])
	}
}

func TestLauncherRestartsInsteadOfResurrectingWhenConfigured(t *testing.T) {
	runner := &recordingRunner{}
	launcher := Launcher{Runner: runner, Env: staticEnv{}}
	err := launcher.Connect(context.Background(), domain.Entry{
		Type:                  domain.EntryProject,
		Name:                  "api",
		SessionName:           "api",
		Path:                  "/tmp/api",
		SessionState:          domain.SessionStateResurrectable,
		RestartOnResurrection: true,
	})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected delete plus recreate, got %#v", runner.calls)
	}
	if runner.calls[0] != "zellij delete-session api" {
		t.Fatalf("expected delete-session first, got %#v", runner.calls)
	}
	if runner.calls[1] != "zellij attach -c api options --default-cwd /tmp/api" {
		t.Fatalf("expected fresh create after delete, got %#v", runner.calls)
	}
}

func TestLauncherCreateSessionPrefersLayoutOverStartup(t *testing.T) {
	runner := &recordingRunner{}
	launcher := Launcher{Runner: runner, Env: staticEnv{}}
	err := launcher.Connect(context.Background(), domain.Entry{SessionName: "api", Path: "/tmp/api", Layout: "compact", Startup: "nvim ."})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if runner.calls[0] != "zellij attach -c api options --default-layout compact --default-cwd /tmp/api" {
		t.Fatalf("unexpected call: %#v", runner.calls)
	}
}

func TestLauncherCreateSessionWithPlainCwdUsesOptions(t *testing.T) {
	runner := &recordingRunner{}
	launcher := Launcher{Runner: runner, Env: staticEnv{}}
	err := launcher.Connect(context.Background(), domain.Entry{SessionName: "api", Path: "/tmp/api"})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if runner.calls[0] != "zellij attach -c api options --default-cwd /tmp/api" {
		t.Fatalf("unexpected call: %#v", runner.calls)
	}
}
