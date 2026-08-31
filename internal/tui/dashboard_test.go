package tui

import (
	"github.com/Is0cre/Akilix/internal/scope"
	"github.com/Is0cre/Akilix/internal/workbookview"
	"strings"
	"testing"
)

func TestDashboardShowsPlaybookReadinessWithoutColor(t *testing.T) {
	o := workbookview.Overview{Name: "lab", ID: "018f0000-0000-7000-8000-000000000000", Status: "open"}
	out := Render(o, scope.Config{Includes: []string{"192.168.1.0/24"}}, false)
	if !strings.Contains(out, "Local network discovery") || !strings.Contains(out, "READY") || strings.Contains(out, "\x1b[") {
		t.Fatalf("dashboard: %q", out)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if len([]rune(line)) != 63 {
			t.Fatalf("unaligned line width %d: %q", len([]rune(line)), line)
		}
	}
}

func TestActionsRenderAsAlignedListWithoutColor(t *testing.T) {
	out := RenderActions(false)
	for _, want := range []string{"Add scope target", "Add exclusion", "Network discovery", "Port discovery", "Open live log", "Leave workbook"} {
		if !strings.Contains(out, want) {
			t.Fatalf("actions missing %q: %q", want, out)
		}
	}
	if strings.Contains(out, "\x1b[") || strings.Contains(out, "[a]") {
		t.Fatalf("actions contain legacy or colored controls: %q", out)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if len([]rune(line)) != 63 {
			t.Fatalf("unaligned action width %d: %q", len([]rune(line)), line)
		}
	}
}
