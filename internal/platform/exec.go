package platform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var ErrCommandNotFound = errors.New("command not found")

type Runner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
	RunInteractive(ctx context.Context, name string, args ...string) error
}

type ExecRunner struct{}

func (ExecRunner) LookPath(file string) (string, error) {
	path, err := exec.LookPath(file)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("%w: %s", ErrCommandNotFound, file)
		}
		return "", err
	}
	return path, nil
}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		message := fmt.Sprintf("run %s %v: %v", name, args, err)
		trimmed := strings.TrimSpace(string(out))
		if trimmed != "" {
			message += ": " + trimmed
		}
		return out, errors.New(message)
	}
	return out, nil
}

func (ExecRunner) RunInteractive(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s %v: %w", name, args, err)
	}
	return nil
}
