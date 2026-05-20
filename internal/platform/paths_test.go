package platform

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigPathForGOOSUnix(t *testing.T) {
	t.Setenv("HOME", "/tmp/test-home")
	path, err := defaultConfigPathForGOOS("linux")
	if err != nil {
		t.Fatalf("defaultConfigPathForGOOS() error = %v", err)
	}
	want := filepath.Join("/tmp/test-home", ".config", "zjsh", "config.kdl")
	if path != want {
		t.Fatalf("expected %q, got %q", want, path)
	}
}

func TestDefaultConfigPathForGOOSWindows(t *testing.T) {
	t.Setenv("APPDATA", filepath.Join("C:", "Users", "tester", "AppData", "Roaming"))
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	path, err := defaultConfigPathForGOOS("windows")
	if err != nil {
		t.Fatalf("defaultConfigPathForGOOS() error = %v", err)
	}
	if !strings.Contains(strings.ToLower(filepath.ToSlash(path)), "/appdata/roaming/zjsh/config.kdl") {
		t.Fatalf("expected Windows config path under AppData, got %q", path)
	}
}
