package service

import (
	"context"
	"path/filepath"
	"sort"

	"github.com/saweima12/zjsh/internal/domain"
	configprovider "github.com/saweima12/zjsh/internal/provider/config"
	"github.com/saweima12/zjsh/internal/provider/zellij"
	"github.com/saweima12/zjsh/internal/provider/zoxide"
)

const (
	projectScore       = 400
	liveSessionScore   = 300
	resurrectableScore = 200
	pathScore          = 100
)

type Aggregator struct {
	Config configprovider.Loader
	Zellij zellij.Provider
	Zoxide zoxide.Provider
}

type ListResult struct {
	Entries []domain.Entry
	Config  domain.Config
}

func (a Aggregator) ListEntries(ctx context.Context, includeResurrectable bool) ([]domain.Entry, error) {
	result, err := a.List(ctx, includeResurrectable)
	if err != nil {
		return nil, err
	}
	return result.Entries, nil
}

func (a Aggregator) List(ctx context.Context, includeResurrectable bool) (ListResult, error) {
	config, err := a.Config.Load(ctx)
	if err != nil {
		return ListResult{}, err
	}
	sessions, err := a.Zellij.ListSessions(ctx)
	if err != nil {
		return ListResult{}, err
	}
	paths, err := a.Zoxide.ListPaths(ctx)
	if err != nil {
		return ListResult{}, err
	}

	entries := make([]domain.Entry, 0, len(config.Projects)+len(sessions)+len(paths))
	order := 0
	for _, project := range config.Projects {
		sessionName := project.Session
		if sessionName == "" {
			sessionName = project.Name
		}
		restartOnResurrection := config.Defaults.RestartOnResurrection
		if project.RestartOnResurrection != nil {
			restartOnResurrection = *project.RestartOnResurrection
		}
		entries = append(entries, domain.Entry{
			Name:                  project.Name,
			Type:                  domain.EntryProject,
			Sources:               []string{"config"},
			Path:                  project.Path,
			SessionName:           sessionName,
			Shell:                 config.Defaults.Shell,
			Startup:               project.Startup,
			Layout:                project.Layout,
			LayoutFile:            project.LayoutFile,
			RestartOnResurrection: restartOnResurrection,
			Score:                 projectScore,
			Order:                 order,
		})
		order++
	}
	for _, session := range sessions {
		if session.State == domain.SessionStateResurrectable && !includeResurrectable {
			continue
		}
		score := liveSessionScore
		if session.State == domain.SessionStateResurrectable {
			score = resurrectableScore
		}
		entries = append(entries, domain.Entry{
			Name:         session.Name,
			Type:         domain.EntrySession,
			Sources:      []string{"zellij"},
			SessionName:  session.Name,
			SessionState: session.State,
			Score:        score,
			Order:        order,
		})
		order++
	}
	for _, path := range paths {
		entries = append(entries, domain.Entry{
			Name:    filepath.Base(path),
			Type:    domain.EntryPath,
			Sources: []string{"zoxide"},
			Path:    path,
			Score:   pathScore,
			Order:   order,
		})
		order++
	}
	return ListResult{Entries: dedupeEntries(entries), Config: config}, nil
}

func dedupeEntries(entries []domain.Entry) []domain.Entry {
	merged := make(map[string]domain.Entry, len(entries))
	aliases := make(map[string]string, len(entries)*3)
	order := make([]string, 0, len(entries))
	for _, entry := range entries {
		key := canonicalKey(entry, aliases, merged)
		existing, ok := merged[key]
		if !ok {
			normalized := normalizeEntry(entry)
			merged[key] = normalized
			registerAliases(aliases, key, normalized)
			order = append(order, key)
			continue
		}
		updated := mergeEntries(existing, entry)
		merged[key] = updated
		registerAliases(aliases, key, updated)
	}
	result := make([]domain.Entry, 0, len(order))
	for _, key := range order {
		result = append(result, merged[key])
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score != result[j].Score {
			return result[i].Score > result[j].Score
		}
		return result[i].Order < result[j].Order
	})
	return result
}

func canonicalKey(entry domain.Entry, aliases map[string]string, merged map[string]domain.Entry) string {
	if entry.Type == domain.EntryProject {
		for _, key := range projectIdentityKeys(entry) {
			if canonical, ok := aliases[key]; ok {
				return canonical
			}
		}
		return primaryKey(entry)
	}

	if entry.Type == domain.EntryPath && entry.Path != "" {
		return "zoxide:" + entry.Path
	}

	for _, key := range identifierKeys(entry) {
		if canonical, ok := aliases[key]; ok {
			return canonical
		}
	}
	return primaryKey(entry)
}

func projectIdentityKeys(entry domain.Entry) []string {
	keys := make([]string, 0, 2)
	if entry.Path != "" {
		keys = append(keys, "path:"+entry.Path)
	}
	if entry.Name != "" {
		keys = append(keys, "name:"+entry.Name)
	}
	return keys
}

func primaryKey(entry domain.Entry) string {
	if entry.Path != "" {
		return "path:" + entry.Path
	}
	if entry.SessionName != "" {
		return "session:" + entry.SessionName
	}
	return "name:" + entry.Name
}

func identifierKeys(entry domain.Entry) []string {
	if entry.Type == domain.EntryPath && entry.Path != "" {
		return []string{"zoxide:" + entry.Path}
	}

	keys := make([]string, 0, 3)
	if entry.Path != "" {
		keys = append(keys, "path:"+entry.Path)
	}
	if entry.SessionName != "" {
		keys = append(keys, "session:"+entry.SessionName)
	}
	if entry.Name != "" {
		keys = append(keys, "name:"+entry.Name)
	}
	return keys
}

func registerAliases(aliases map[string]string, canonical string, entry domain.Entry) {
	for _, key := range identifierKeys(entry) {
		aliases[key] = canonical
	}
}

func normalizeEntry(entry domain.Entry) domain.Entry {
	entry.Sources = uniqueSorted(entry.Sources)
	return entry
}

func mergeEntries(existing, incoming domain.Entry) domain.Entry {
	primary := existing
	secondary := incoming
	if incoming.Score > existing.Score {
		primary, secondary = incoming, existing
	}
	merged := primary
	merged.Sources = uniqueSorted(append(existing.Sources, incoming.Sources...))
	if merged.Path == "" {
		merged.Path = secondary.Path
	}
	if merged.SessionName == "" {
		merged.SessionName = secondary.SessionName
	}
	if merged.SessionState == domain.SessionStateNone {
		merged.SessionState = secondary.SessionState
	}
	if merged.Startup == "" {
		merged.Startup = secondary.Startup
	}
	if merged.Shell == "" {
		merged.Shell = secondary.Shell
	}
	if merged.Layout == "" {
		merged.Layout = secondary.Layout
	}
	if merged.LayoutFile == "" {
		merged.LayoutFile = secondary.LayoutFile
	}
	merged.RestartOnResurrection = merged.RestartOnResurrection || secondary.RestartOnResurrection
	if merged.Name == "" {
		merged.Name = secondary.Name
	}
	if merged.Score < secondary.Score {
		merged.Score = secondary.Score
	}
	if secondary.Order < merged.Order {
		merged.Order = secondary.Order
	}
	return merged
}

func uniqueSorted(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
