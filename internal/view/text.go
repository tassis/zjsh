package view

import (
	"io"

	"zjsh/internal/domain"
)

func WriteTable(w io.Writer, entries []domain.Entry) error {
	for _, entry := range entries {
		if _, err := io.WriteString(w, selectorLabel(entry)+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func selectorLabel(entry domain.Entry) string {
	return selectorIcon(entry) + " " + selectorValue(entry)
}

func selectorValue(entry domain.Entry) string {
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

func selectorIcon(entry domain.Entry) string {
	switch entry.Type {
	case domain.EntryProject:
		return "◆"
	case domain.EntryPath:
		return "→"
	case domain.EntrySession:
		if entry.SessionState == domain.SessionStateResurrectable {
			return "↺"
		}
		return "●"
	default:
		if entry.SessionState == domain.SessionStateResurrectable {
			return "↺"
		}
		return "•"
	}
}
