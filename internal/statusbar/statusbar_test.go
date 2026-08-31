package statusbar

import (
	"testing"
	"time"

	"github.com/Is0cre/Akilix/internal/activity"
	"github.com/Is0cre/Akilix/internal/scope"
	"github.com/Is0cre/Akilix/internal/workbook"
)

func TestMetricsAreDerivedFromCanonicalLocalState(t *testing.T) {
	root, runtime := t.TempDir(), t.TempDir()
	m, err := workbook.Create(root, "case", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := scope.Add(root+"/case", "192.168.2.0/24", false); err != nil {
		t.Fatal(err)
	}
	if err := Activate(runtime, root, "case"); err != nil {
		t.Fatal(err)
	}
	if err := activity.Append(root+"/case", activity.Event{Schema: activity.Schema, InvocationID: "one", WorkbookID: m.ID, Timestamp: time.Now(), Phase: "STARTED", Executor: "container", Tool: "naabu"}); err != nil {
		t.Fatal(err)
	}
	got, err := Current(runtime)
	if err != nil || got.Workbook != "case" || got.Scope != "192.168.2.0/24" || got.Running != 1 || got.Failed != 0 {
		t.Fatalf("metrics=%+v err=%v", got, err)
	}
}
