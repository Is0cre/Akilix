package invocation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/pensuse/pensuse/internal/activity"
)

func TestRunRecordsSuccessAndFailure(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PENSUSE_TEST_SECRET", "do-not-record")
	now := func() time.Time { return time.Unix(10, 0) }
	startedID := ""
	r, err := RunWithOptions(context.Background(), root, "88888888-8888-7888-8888-888888888888", []string{"sh", "-c", "[ -z \"$PENSUSE_TEST_SECRET\" ] || exit 9; printf out; printf err >&2; printf generated > '" + filepath.Join(root, "tool-output", "generated.txt") + "'"}, now, Options{ScopeResult: "DENY", ScopeTarget: "10.0.0.1", ScopeOverride: true, OnStarted: func(id string) { startedID = id }})
	if err != nil || r.Status != "complete" || r.ExitCode != 0 {
		t.Fatalf("success: %+v %v", r, err)
	}
	if startedID != r.ID {
		t.Fatalf("start callback id=%q record=%q", startedID, r.ID)
	}
	if r.ScopeResult != "DENY" || r.ScopeTarget != "10.0.0.1" || !r.ScopeOverride {
		t.Fatalf("scope provenance missing: %+v", r)
	}
	if r.WorkingDirectory == "" || r.Environment == nil {
		t.Fatalf("execution environment provenance missing: %+v", r)
	}
	if _, found := r.Environment["PENSUSE_TEST_SECRET"]; found {
		t.Fatal("unapproved environment variable recorded")
	}
	if len(r.GeneratedFiles) != 1 || r.GeneratedFiles[0] != "tool-output/generated.txt" {
		t.Fatalf("generated artifact provenance missing: %+v", r.GeneratedFiles)
	}
	out, _ := os.ReadFile(filepath.Join(root, r.Stdout))
	if string(out) != "out" {
		t.Fatalf("stdout %q", out)
	}
	r, err = Run(context.Background(), root, "88888888-8888-7888-8888-888888888888", []string{"sh", "-c", "exit 7"}, now)
	if err == nil || r.Status != "failed" || r.ExitCode != 7 {
		t.Fatalf("failure: %+v %v", r, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".pensuse", "manifest.jsonl")); err != nil {
		t.Fatal(err)
	}
	events, err := activity.List(root)
	if err != nil || len(events) != 4 || events[0].Phase != "STARTED" || events[1].Phase != "COMPLETED" || events[2].Phase != "STARTED" || events[3].Phase != "FAILED" {
		t.Fatalf("lifecycle events=%+v err=%v", events, err)
	}
}

func TestAppendRecordSerializesConcurrentWriters(t *testing.T) {
	root := t.TempDir()
	const writers = 32
	var wg sync.WaitGroup
	errs := make(chan error, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			record := Record{
				Schema: Schema, ID: fmt.Sprintf("77777777-7777-7777-8777-%012x", i),
				WorkbookID: "88888888-8888-7888-8888-888888888888",
				Started:    time.Unix(1, 0), Ended: time.Unix(2, 0), Executor: "native",
				Executable: "/bin/true", Arguments: []string{"true"}, ExitCode: 0,
				Status: "complete", Stdout: "tool-output/a.stdout", Stderr: "tool-output/a.stderr",
			}
			errs <- appendRecord(root, record)
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	records, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != writers {
		t.Fatalf("got %d invocation records, want %d", len(records), writers)
	}
}
