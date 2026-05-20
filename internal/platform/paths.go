package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func DefaultConfigPath() (string, error) {
	return defaultConfigPathForGOOS(runtime.GOOS)
}

func defaultConfigPathForGOOS(goos string) (string, error) {
	if goos == "windows" {
		configDir := os.Getenv("APPDATA")
		if configDir == "" {
			var err error
			configDir, err = os.UserConfigDir()
			if err != nil {
				return "", err
			}
		}
		return filepath.Join(configDir, "zjsh", "config.kdl"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "zjsh", "config.kdl"), nil
}

func ExpandHome(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	if !strings.HasPrefix(path, "~/") {
		return "", fmt.Errorf("unsupported home path: %s", path)
	}
	return filepath.Join(home, path[2:]), nil
}
