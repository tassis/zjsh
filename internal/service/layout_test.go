package service

import (
	"strings"
	"testing"
)

func TestBuildStartupLayout(t *testing.T) {
	layout := BuildStartupLayout("bash", "/tmp/api", "nvim .")
	if !strings.Contains(layout, `command="bash"`) {
		t.Fatalf("expected shell in layout, got %q", layout)
	}
	if !strings.Contains(layout, `args "-lc" "nvim ."`) {
		t.Fatalf("expected startup args in layout, got %q", layout)
	}
	if !strings.Contains(layout, `cwd "/tmp/api"`) {
		t.Fatalf("expected cwd in layout, got %q", layout)
	}
}
