package service

import (
	"fmt"
	"strings"

	"github.com/tassis/zjsh/internal/domain"
)

func ResolveMacro(macros []domain.Macro, target string) (domain.Macro, error) {
	target = strings.TrimSpace(target)
	for _, macro := range macros {
		if macro.Name == target {
			return macro, nil
		}
	}
	return domain.Macro{}, fmt.Errorf("%w: %s", ErrTargetNotFound, target)
}
