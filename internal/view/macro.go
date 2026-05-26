package view

import (
	"encoding/json"
	"io"

	"github.com/tassis/zjsh/internal/domain"
)

func WriteMacroPlain(w io.Writer, macros []domain.Macro) error {
	for _, macro := range macros {
		if _, err := io.WriteString(w, macro.Name+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func WriteMacroLabels(w io.Writer, macros []domain.Macro, icons domain.Icons) error {
	icon := macroIcon(icons)
	for _, macro := range macros {
		if _, err := io.WriteString(w, icon+" "+macro.Name+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func WriteMacroJSON(w io.Writer, macros []domain.Macro) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(macros)
}

func macroIcon(icons domain.Icons) string {
	if icons.Macro != "" {
		return icons.Macro
	}
	return domain.DefaultIcons().Macro
}
