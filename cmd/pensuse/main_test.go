package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/pensuse/pensuse/internal/version"
)

func TestVersionJSONShape(t *testing.T) {
	b, err := json.Marshal(version.Current())
	if err != nil { t.Fatal(err) }
	var got map[string]string
	if err := json.Unmarshal(b, &got); err != nil { t.Fatal(err) }
	for _, key := range []string{"name", "version", "base", "architecture"} {
		if got[key] == "" { t.Fatalf("missing %s", key) }
	}
}

func TestRunVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"version"}, &out, &errOut); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut.String())
	}
	if out.String() == "" || errOut.Len() != 0 {
		t.Fatalf("unexpected output: stdout=%q stderr=%q", out.String(), errOut.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"status"}, &out, &errOut); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}
