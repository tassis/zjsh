package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/saweima12/zjsh/internal/domain"
	"github.com/saweima12/zjsh/internal/platform"
)

type Launcher struct {
	Runner platform.Runner
	Env    platform.Env
}

func (l Launcher) Connect(ctx context.Context, entry domain.Entry) error {
	sessionName := effectiveSessionName(entry)
	if sessionName == "" {
		return fmt.Errorf("entry is missing a session name")
	}
	if entry.SessionState == domain.SessionStateLive {
		name, args := liveSessionArgs(sessionName, l.inZellij())
		return l.run(ctx, name, args...)
	}
	if entry.SessionState == domain.SessionStateResurrectable {
		if entry.RestartOnResurrection {
			if err := l.forgetSession(ctx, sessionName); err != nil {
				return err
			}
			entry.SessionState = domain.SessionStateNone
			name, args := createSessionArgs(entry, l.inZellij())
			return l.run(ctx, name, args...)
		}
		name, args := liveSessionArgs(sessionName, l.inZellij())
		return l.run(ctx, name, args...)
	}
	name, args := createSessionArgs(entry, l.inZellij())
	return l.run(ctx, name, args...)
}

func (l Launcher) forgetSession(ctx context.Context, sessionName string) error {
	_, err := l.Runner.Run(ctx, "zellij", "delete-session", sessionName)
	return err
}

func (l Launcher) inZellij() bool {
	if l.Env == nil {
		return false
	}
	return l.Env.InZellij()
}

func (l Launcher) run(ctx context.Context, name string, args ...string) error {
	return l.Runner.RunInteractive(ctx, name, args...)
}

func liveSessionArgs(sessionName string, inZellij bool) (string, []string) {
	if inZellij {
		return "zellij", []string{"action", "switch-session", sessionName}
	}
	return "zellij", []string{"attach", sessionName}
}

func createSessionArgs(entry domain.Entry, inZellij bool) (string, []string) {
	if inZellij {
		return createSessionArgsInsideZellij(entry)
	}
	return createSessionArgsOutsideZellij(entry)
}

func createSessionArgsInsideZellij(entry domain.Entry) (string, []string) {
	sessionName := effectiveSessionName(entry)
	path := entry.Path
	args := []string{"action", "switch-session", sessionName}
	if entry.LayoutFile != "" {
		args = append(args, "--layout", entry.LayoutFile)
		if path != "" {
			args = append(args, "--cwd", path)
		}
		return "zellij", args
	}
	if entry.Layout != "" {
		args = append(args, "--layout", entry.Layout)
		if path != "" {
			args = append(args, "--cwd", path)
		}
		return "zellij", args
	}
	if path != "" {
		args = append(args, "--cwd", path)
	}
	return "zellij", args
}

func createSessionArgsOutsideZellij(entry domain.Entry) (string, []string) {
	sessionName := effectiveSessionName(entry)
	path := entry.Path
	options := []string{"attach", "-c", sessionName, "options"}
	if entry.LayoutFile != "" {
		options = append(options, "--default-layout", entry.LayoutFile)
		if path != "" {
			options = append(options, "--default-cwd", path)
		}
		return "zellij", options
	}
	if entry.Layout != "" {
		options = append(options, "--default-layout", entry.Layout)
		if path != "" {
			options = append(options, "--default-cwd", path)
		}
		return "zellij", options
	}
	if path != "" {
		options = append(options, "--default-cwd", path)
	}
	return "zellij", options
}

func effectiveSessionName(entry domain.Entry) string {
	if entry.SessionName != "" {
		return entry.SessionName
	}
	if entry.Path != "" {
		return filepath.Base(entry.Path)
	}
	return entry.Name
}
