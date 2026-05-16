package view

import (
	"encoding/json"
	"io"

	"github.com/saweima12/zjsh/internal/domain"
)

func WriteJSON(w io.Writer, entries []domain.Entry) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
