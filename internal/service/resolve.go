package service

import (
	"errors"
	"fmt"

	"github.com/saweima12/zjsh/internal/domain"
)

var ErrTargetNotFound = errors.New("target not found")

func ResolveExact(entries []domain.Entry, target string) (domain.Entry, error) {
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
		if entry.Path == target {
			return entry, true
		}
	}
	return domain.Entry{}, false
}
