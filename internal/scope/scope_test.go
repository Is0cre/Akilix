package scope

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Is0cre/Akilix/internal/journal"
)

func TestScopeEvaluation(t *testing.T) {
	c := Config{Includes: []string{"10.0.0.0/8", "*.example.org", "https://portal.example.net/app"}, Excludes: []string{"10.0.5.0/24", "secret.example.org"}}
	cases := map[string]Result{"10.1.2.3": Allow, "10.0.5.2": Deny, "https://app.example.org/x": Allow, "secret.example.org": Deny, "https://portal.example.net/other": Allow, "other.test": Unknown}
	for target, want := range cases {
		if got := Evaluate(c, target); got != want {
			t.Errorf("%s: got %s want %s", target, got, want)
		}
	}
}

func TestExclusionMayOverrideInclude(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "scope.yaml"), []byte("version: 1\ninclude:\n  - 'Example.COM'\nexclude:\n  - 'example.com'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := Evaluate(c, "example.com"); got != Deny {
		t.Fatalf("exclusion did not override include: %s", got)
	}
}

func TestDecisionIncludesMatchingRule(t *testing.T) {
	c := Config{Includes: []string{"10.0.0.0/8"}, Excludes: []string{"10.0.5.0/24"}}
	if got := EvaluateDecision(c, "10.0.5.2"); got.Result != Deny || got.Rule != "10.0.5.0/24" {
		t.Fatalf("deny decision: %+v", got)
	}
	if got := EvaluateDecision(c, "10.1.2.3"); got.Result != Allow || got.Rule != "10.0.0.0/8" {
		t.Fatalf("allow decision: %+v", got)
	}
	if got := EvaluateDecision(c, "other.test"); got.Result != Unknown || got.Rule != "" {
		t.Fatalf("unknown decision: %+v", got)
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

func TestScopeRejectsWhitespaceAndCanonicalizesIP(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "scope.yaml"), []byte("version: 1\ninclude:\n  - 2001:0DB8::/32\nexclude:\n  []\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(root)
	if err != nil || len(c.Includes) != 1 || c.Includes[0] != "2001:db8::/32" {
		t.Fatalf("canonicalization: %+v %v", c, err)
	}
	if err := Add(root, "bad target", false); err == nil {
		t.Fatal("whitespace target accepted")
	}
}

func TestNormalizeIPTargetUsesNetipAndRejectsHostnames(t *testing.T) {
	for input, want := range map[string]string{"192.168.2.9/24": "192.168.2.0/24", "10.0.0.5": "10.0.0.5", "2001:db8::1/64": "2001:db8::/64"} {
		got, err := NormalizeIPTarget(input)
		if err != nil || got != want {
			t.Fatalf("input=%q got=%q want=%q err=%v", input, got, want, err)
		}
	}
	if _, err := NormalizeIPTarget("example.org"); err == nil {
		t.Fatal("hostname accepted by IP-only modal validator")
	}
}

func TestAddRecordedWritesRequestThenCompletion(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "scope.yaml"), []byte("version: 1\ninclude:\n  []\nexclude:\n  []\n"), 0600); err != nil {
		t.Fatal(err)
	}
	log, err := journal.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := AddRecorded(root, "workbook-id", "192.168.2.9/24", false, log, time.Now())
	if err != nil || got != "192.168.2.0/24" {
		t.Fatalf("got=%q err=%v", got, err)
	}
	config, err := Load(root)
	if err != nil || len(config.Includes) != 1 || config.Includes[0] != got {
		t.Fatalf("config=%+v err=%v", config, err)
	}
	file, err := os.Open(log.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var events []journal.Event
	for scanner.Scan() {
		var event journal.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	if len(events) != 2 || events[0].Event != "SCOPE_TARGET_ADD_REQUESTED" || events[1].Event != "SCOPE_TARGET_ADDED" || events[1].Payload["value"] != got {
		t.Fatalf("events=%+v", events)
	}
}
