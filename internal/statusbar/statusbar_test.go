package statusbar

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Is0cre/Akilix/internal/activity"
	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/scope"
	"github.com/Is0cre/Akilix/internal/workbook"
)

func TestExplicitActiveContextCanBeReadAndRemoved(t *testing.T) {
	root, runtime := t.TempDir(), t.TempDir()
	if _, err := workbook.Create(root, "case", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := Activate(runtime, root, "case"); err != nil {
		t.Fatal(err)
	}
	state, err := Active(runtime)
	if err != nil || state.Workbook != "case" || state.Root != root {
		t.Fatalf("state=%+v err=%v", state, err)
	}
	if err := Deactivate(runtime); err != nil {
		t.Fatal(err)
	}
	if err := Deactivate(runtime); err != nil {
		t.Fatalf("deactivation is not idempotent: %v", err)
	}
	if _, err := Active(runtime); !os.IsNotExist(err) {
		t.Fatalf("active after deactivation: %v", err)
	}
}

func TestActiveRejectsSymlinkState(t *testing.T) {
	runtime := t.TempDir()
	dir := filepath.Join(runtime, "akilix")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("elsewhere", filepath.Join(dir, "active-workbook.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := Active(runtime); err == nil {
		t.Fatal("symlink state accepted")
	}
}

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
	log, err := journal.Open(root + "/case")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"HOST_DISCOVERED", "HOST_DISCOVERED", "PORT_FOUND", "PORT_DROPPED_OUT_OF_SCOPE"} {
		event, eventErr := journal.NewEvent(name, "RECON", map[string]any{"value": "192.168.2.1"}, time.Now())
		if eventErr != nil {
			t.Fatal(eventErr)
		}
		if err := log.Append(event); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Current(runtime)
	if err != nil || got.Workbook != "case" || got.Scope != "192.168.2.0/24" || got.Running != 1 || got.Failed != 0 || got.Hosts != 2 || got.Ports != 1 || got.Dropped != 1 {
		t.Fatalf("metrics=%+v err=%v", got, err)
	}
}

func TestBlocksRenderPassiveDiscoveryTelemetry(t *testing.T) {
	blocks := Blocks(Metrics{Workbook: "case", Scope: "192.0.2.0/24", Hosts: 7, Ports: 19, Dropped: 2})
	if len(blocks) != 5 || blocks[2].FullText != " 📡 HOSTS 7 · PORTS 19 · DROPPED 2 " || blocks[2].Color != "#72b7c9" {
		t.Fatalf("unexpected discovery block: %+v", blocks)
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
