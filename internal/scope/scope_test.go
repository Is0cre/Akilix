package scope

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScopeEvaluation(t *testing.T) {
	c := Config{Includes: []string{"10.0.0.0/8", "*.example.org"}, Excludes: []string{"10.0.5.0/24", "secret.example.org"}}
	cases := map[string]Result{"10.1.2.3": Allow, "10.0.5.2": Deny, "https://app.example.org/x": Allow, "secret.example.org": Deny, "other.test": Unknown}
	for target, want := range cases {
		if got := Evaluate(c, target); got != want {
			t.Errorf("%s: got %s want %s", target, got, want)
		}
	}
}
func TestLoadSave(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "scope.yaml"), []byte("version: 1\ninclude:\n  - 192.0.2.0/24\nexclude:\n  - 192.0.2.10\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(root)
	if err != nil || len(c.Includes) != 1 || len(c.Excludes) != 1 {
		t.Fatalf("load: %+v %v", c, err)
	}
	if err := Save(root, c); err != nil {
		t.Fatal(err)
	}
	if err := Remove(root, "192.0.2.0/24", false); err != nil {
		t.Fatal(err)
	}
	if err := Remove(root, "missing", false); err == nil {
		t.Fatal("missing scope removal accepted")
	}
}
