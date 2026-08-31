package workbookview

import (
	"strings"
	"testing"
)

func TestRenderProvidesNavigablePassiveWorkspace(t *testing.T) {
	overview := Overview{
		Name: "case-1", Status: "open", ID: "workbook-id", Created: "2026-08-31T12:00:00Z",
		Root: "/cases/case-1", Sections: []Section{{Name: "evidence", Path: "/cases/case-1/evidence"}},
	}
	got := Render(overview)
	for _, want := range []string{
		"Akilix / WORKBOOK", "case-1  [OPEN]", "WORKSPACE", "CAPTURE POLICY",
		"evidence", `akilix workbook path case-1 evidence`,
		"Read-only view · no network activity · original evidence stays immutable",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered workspace missing %q:\n%s", want, got)
		}
	}
}
