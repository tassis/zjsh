package version

import "testing"

func TestString(t *testing.T) {
	old := Version
	Version = "v1.2.3"
	t.Cleanup(func() { Version = old })

	if got := String(); got != "zjsh v1.2.3" {
		t.Fatalf("String() = %q", got)
	}
}
