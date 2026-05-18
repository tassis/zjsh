package zoxide

import (
	"context"
	"errors"
	"testing"

	"github.com/saweima12/zjsh/internal/platform"
)

type fakeRunner struct {
	lookPathErr error
	runErr      error
	output      string
}

func (f fakeRunner) LookPath(string) (string, error) {
	if f.lookPathErr != nil {
		return "", f.lookPathErr
	}
	return "/usr/bin/zoxide", nil
}

func (f fakeRunner) Run(context.Context, string, ...string) ([]byte, error) {
	return []byte(f.output), f.runErr
}

func (f fakeRunner) RunInteractive(context.Context, string, ...string) error {
	return nil
}

func TestParsePaths(t *testing.T) {
	paths := ParsePaths("/tmp/api\n\n/tmp/api\n/tmp/blog\n")
	if len(paths) != 2 {
		t.Fatalf("expected 2 unique paths, got %d", len(paths))
	}
	if paths[0] != "/tmp/api" || paths[1] != "/tmp/blog" {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}

func TestListPathsSkipsMissingZoxide(t *testing.T) {
	paths, err := Provider{Runner: fakeRunner{lookPathErr: platform.ErrCommandNotFound}}.ListPaths(context.Background())
	if err != nil {
		t.Fatalf("ListPaths() error = %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no paths, got %#v", paths)
	}
}

func TestListPathsSkipsLookPathFailure(t *testing.T) {
	paths, err := Provider{Runner: fakeRunner{lookPathErr: errors.New("permission denied")}}.ListPaths(context.Background())
	if err != nil {
		t.Fatalf("ListPaths() error = %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no paths, got %#v", paths)
	}
}

func TestListPathsSkipsQueryFailure(t *testing.T) {
	paths, err := Provider{Runner: fakeRunner{runErr: errors.New("query failed")}}.ListPaths(context.Background())
	if err != nil {
		t.Fatalf("ListPaths() error = %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no paths, got %#v", paths)
	}
}
