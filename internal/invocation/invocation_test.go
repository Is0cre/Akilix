package invocation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunRecordsSuccessAndFailure(t *testing.T) {
	root := t.TempDir()
	now := func() time.Time { return time.Unix(10, 0) }
	r, err := RunWithOptions(context.Background(), root, "wb-1", []string{"sh", "-c", "printf out; printf err >&2"}, now, Options{ScopeResult: "ALLOW", ScopeOverride: true})
	if err != nil || r.Status != "complete" || r.ExitCode != 0 {
		t.Fatalf("success: %+v %v", r, err)
	}
	if r.ScopeResult != "ALLOW" || !r.ScopeOverride {
		t.Fatalf("scope provenance missing: %+v", r)
	}
	out, _ := os.ReadFile(filepath.Join(root, r.Stdout))
	if string(out) != "out" {
		t.Fatalf("stdout %q", out)
	}
	r, err = Run(context.Background(), root, "wb-1", []string{"sh", "-c", "exit 7"}, now)
	if err == nil || r.Status != "failed" || r.ExitCode != 7 {
		t.Fatalf("failure: %+v %v", r, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".pensuse", "manifest.jsonl")); err != nil {
		t.Fatal(err)
	}
}
