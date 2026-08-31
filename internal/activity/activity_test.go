package activity

import (
	"testing"
	"time"
)

func TestAppendAndListLifecycle(t *testing.T) {
	root := t.TempDir()
	now := time.Now().UTC()
	for _, phase := range []string{"STARTED", "COMPLETED"} {
		if err := Append(root, Event{Schema: Schema, InvocationID: "inv", WorkbookID: "wb", Timestamp: now, Phase: phase, Executor: "container", Tool: "httpx"}); err != nil {
			t.Fatal(err)
		}
	}
	events, err := List(root)
	if err != nil || len(events) != 2 || events[0].Phase != "STARTED" || events[1].Phase != "COMPLETED" {
		t.Fatalf("events=%+v err=%v", events, err)
	}
}
