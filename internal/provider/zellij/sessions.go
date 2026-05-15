package zellij

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"zjsh/internal/domain"
	"zjsh/internal/platform"
)

type Session struct {
	Name  string
	State domain.SessionState
}

type Provider struct {
	Runner platform.Runner
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func (p Provider) ListSessions(ctx context.Context) ([]Session, error) {
	if _, err := p.Runner.LookPath("zellij"); err != nil {
		if errors.Is(err, platform.ErrCommandNotFound) {
			return nil, nil
		}
		return nil, err
	}
	out, err := p.Runner.Run(ctx, "zellij", "list-sessions", "-n")
	if err != nil {
		if isNoActiveSessions(out, err) {
			return nil, nil
		}
		return nil, err
	}
	return ParseSessions(string(out)), nil
}

func isNoActiveSessions(out []byte, err error) bool {
	message := string(out)
	if err != nil {
		message += "\n" + err.Error()
	}
	return strings.Contains(message, "No active zellij sessions found.")
}

func ParseSessions(input string) []Session {
	var sessions []Session
	seen := map[string]struct{}{}
	for _, raw := range strings.Split(input, "\n") {
		line := strings.TrimSpace(stripANSI(raw))
		if line == "" || isHeaderLine(line) {
			continue
		}
		name := parseName(line)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		state := domain.SessionStateLive
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "[EXITED]") || strings.Contains(upper, "[RESURRECT]") || strings.Contains(upper, "EXITED") {
			state = domain.SessionStateResurrectable
		}
		sessions = append(sessions, Session{Name: name, State: state})
	}
	return sessions
}

func stripANSI(value string) string {
	return ansiPattern.ReplaceAllString(value, "")
}

func isHeaderLine(line string) bool {
	upper := strings.ToUpper(line)
	return strings.Contains(upper, "SESSION") && strings.Contains(upper, "NAME")
}

func parseName(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	for _, field := range fields {
		if strings.HasPrefix(field, "[") && strings.HasSuffix(field, "]") {
			continue
		}
		if strings.EqualFold(field, "EXITED") || strings.EqualFold(field, "LIVE") {
			continue
		}
		return strings.Trim(field, "*")
	}
	return ""
}
