package service

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/tassis/zjsh/internal/domain"
)

var ErrTargetNotFound = errors.New("target not found")

func ResolveExact(entries []domain.Entry, target string) (domain.Entry, error) {
	if entry, ok := findCurrentDir(entries, target); ok {
		return entry, nil
	}
	if entry, ok := findProjectName(entries, target); ok {
		return entry, nil
	}
	if entry, ok := findSessionName(entries, target); ok {
		return entry, nil
	}
	if entry, ok := findPath(entries, target); ok {
		return entry, nil
	}
	return domain.Entry{}, fmt.Errorf("%w: %s", ErrTargetNotFound, target)
}

func findCurrentDir(entries []domain.Entry, target string) (domain.Entry, bool) {
	if target != "." {
		return domain.Entry{}, false
	}
	for _, entry := range entries {
		if entry.CurrentDir {
			return entry, true
		}
	}
	return domain.Entry{}, false
}

func findProjectName(entries []domain.Entry, target string) (domain.Entry, bool) {
	for _, entry := range entries {
		if entry.Type == domain.EntryProject && entry.Name == target {
			return entry, true
		}
	}
	return domain.Entry{}, false
}

func findSessionName(entries []domain.Entry, target string) (domain.Entry, bool) {
	for _, entry := range entries {
		if entry.SessionName == target {
			return entry, true
		}
	}
	return domain.Entry{}, false
}

func findPath(entries []domain.Entry, target string) (domain.Entry, bool) {
	for _, entry := range entries {
		if entry.Type == domain.EntryPath && entry.Path == target {
			return entry, true
		}
	}
	for _, entry := range entries {
		if entry.Path == target {
			return entry, true
		}
	}
	return domain.Entry{}, false
}

func PrepareConnectEntry(entry domain.Entry, entries []domain.Entry) domain.Entry {
	if entry.Type != domain.EntryPath || entry.SessionName != "" || entry.Path == "" {
		return entry
	}
	entry.SessionName = zoxideSessionName(entry.Path, entries)
	return entry
}

func zoxideSessionName(path string, entries []domain.Entry) string {
	desired := filepath.Base(path)
	if desired == "." || desired == string(filepath.Separator) || desired == "" {
		desired = "zoxide"
	}
	if !sessionNameReserved(desired, entries) {
		return desired
	}
	return desired + "-" + shortPathHash(path)
}

func sessionNameReserved(name string, entries []domain.Entry) bool {
	for _, entry := range entries {
		if entry.Type == domain.EntryPath {
			continue
		}
		if entry.SessionName == name {
			return true
		}
	}
	return false
}

func shortPathHash(path string) string {
	sum := sha1.Sum([]byte(path))
	return hex.EncodeToString(sum[:])[:8]
}
