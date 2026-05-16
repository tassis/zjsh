package zoxide

import (
	"context"
	"errors"
	"strings"

	"github.com/saweima12/zjsh/internal/platform"
)

type Provider struct {
	Runner platform.Runner
}

func (p Provider) ListPaths(ctx context.Context) ([]string, error) {
	if _, err := p.Runner.LookPath("zoxide"); err != nil {
		if errors.Is(err, platform.ErrCommandNotFound) {
			return nil, nil
		}
		return nil, err
	}
	out, err := p.Runner.Run(ctx, "zoxide", "query", "-l")
	if err != nil {
		return nil, err
	}
	return ParsePaths(string(out)), nil
}

func ParsePaths(input string) []string {
	var paths []string
	seen := map[string]struct{}{}
	for _, raw := range strings.Split(input, "\n") {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	return paths
}
