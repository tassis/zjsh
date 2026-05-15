package zoxide

import "testing"

func TestParsePaths(t *testing.T) {
	paths := ParsePaths("/tmp/api\n\n/tmp/api\n/tmp/blog\n")
	if len(paths) != 2 {
		t.Fatalf("expected 2 unique paths, got %d", len(paths))
	}
	if paths[0] != "/tmp/api" || paths[1] != "/tmp/blog" {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}
