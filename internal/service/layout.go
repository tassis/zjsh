package service

import (
	"fmt"
	"strings"
)

func BuildStartupLayout(shell, cwd, startup string) string {
	if shell == "" {
		shell = "sh"
	}
	return fmt.Sprintf(
		"layout { pane command=%q { args %q %q cwd %q } }",
		escapeKDLString(shell),
		escapeKDLString("-lc"),
		escapeKDLString(startup),
		escapeKDLString(cwd),
	)
}

func escapeKDLString(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return replacer.Replace(value)
}
