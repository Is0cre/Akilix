package workbookview

import (
	"testing"
	"time"

	"github.com/Is0cre/Akilix/internal/workbook"
)

func TestBuildAndSafeSectionPath(t *testing.T) {
	root := t.TempDir()
	if _, err := workbook.Create(root, "case-1", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	overview, err := Build(root, "case-1")
	if err != nil {
		t.Fatal(err)
	}
	if overview.Name != "case-1" || overview.Evidence != 0 || len(overview.Sections) != 10 {
		t.Fatalf("unexpected overview: %+v", overview)
	}
	if _, err := Path(root, "case-1", "original-evidence"); err != nil {
		t.Fatal(err)
	}
	if _, err := Path(root, "case-1", "../../escape"); err == nil {
		t.Fatal("accepted unknown traversal section")
	}
}
