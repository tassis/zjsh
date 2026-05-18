package view

import (
	"io"

	"github.com/saweima12/zjsh/internal/domain"
)

func WritePlain(w io.Writer, entries []domain.Entry) error {
	for _, entry := range entries {
		if _, err := io.WriteString(w, selectorValue(entry)+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func WriteLabels(w io.Writer, entries []domain.Entry, icons domain.Icons) error {
	for _, entry := range entries {
		if _, err := io.WriteString(w, selectorLabel(entry, icons)+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func selectorLabel(entry domain.Entry, icons domain.Icons) string {
	return selectorIcon(entry, icons) + " " + selectorValue(entry)
}

func selectorValue(entry domain.Entry) string {
	if entry.CurrentDir {
		return "."
	}
	if entry.Type == domain.EntryPath && entry.Path != "" {
		return entry.Path
	}
	if entry.Name != "" {
		return entry.Name
	}
	if entry.SessionName != "" {
		return entry.SessionName
	}
	return entry.Path
}

func selectorIcon(entry domain.Entry, icons domain.Icons) string {
	switch entry.Type {
	case domain.EntryProject:
		return icons.Project
	case domain.EntryPath:
		return icons.Path
	case domain.EntrySession:
		if entry.SessionState == domain.SessionStateResurrectable {
			return icons.Resurrectable
		}
		return icons.Session
	default:
		if entry.SessionState == domain.SessionStateResurrectable {
			return icons.Resurrectable
		}
		return "•"
	}
}
