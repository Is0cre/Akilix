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

func TestTimeBlocksUseLocalCalendarAndZone(t *testing.T) {
	zone := time.FixedZone("EET", 2*60*60)
	blocks := TimeBlocks(time.Date(2026, time.August, 31, 21, 7, 0, 0, zone))
	if len(blocks) != 2 || blocks[0].FullText != " 📅 Mon 31 Aug " || blocks[1].FullText != " 🕒 21:07 EET " {
		t.Fatalf("unexpected time blocks: %+v", blocks)
	}
}

func TestScopeRibbonDistinguishesUndeclaredAndActive(t *testing.T) {
	undeclared := Blocks(Metrics{Workbook: "case", Scope: "UNDECLARED"})
	active := Blocks(Metrics{Workbook: "case", Scope: "192.0.2.0/24"})
	if undeclared[1].FullText != " 🚨 SCOPE UNDECLARED " || active[1].FullText != " 🎯 SCOPE ACTIVE · 192.0.2.0/24 " || active[1].Color != "#87ff00" {
		t.Fatalf("undeclared=%+v active=%+v", undeclared[1], active[1])
	}
}
