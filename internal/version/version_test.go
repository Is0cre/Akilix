package version

import "testing"

func TestCurrent(t *testing.T) {
	got := Current()
	if got.Name != Name || got.Version != Version || got.Base != Base || got.Architecture == "" {
		t.Fatalf("unexpected version info: %+v", got)
	}
}
