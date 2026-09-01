package workbookview

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Is0cre/Akilix/internal/acquire"
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
	if overview.Name != "case-1" || overview.Evidence != 0 || len(overview.Sections) != 11 {
		t.Fatalf("unexpected overview: %+v", overview)
	}
	if _, err := Path(root, "case-1", "original-evidence"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "case-1", "hardware"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := Path(root, "case-1", "hardware"); err != nil {
		t.Fatal(err)
	}
	if _, err := Path(root, "case-1", "../../escape"); err == nil {
		t.Fatal("accepted unknown traversal section")
	}
}

func TestBuildSurfacesAcquisitionRecoveryState(t *testing.T) {
	root := t.TempDir()
	metadata, err := workbook.Create(root, "case-1", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	workbookRoot := filepath.Join(root, "case-1")
	record := acquire.ImageRecord{Schema: acquire.ImageSchema, OperationID: "018f1f1e-7b8a-7000-8000-000000000001", WorkbookID: metadata.ID, Phase: "REQUESTED", RecordedAt: time.Now(), Source: "/dev/sdb", Destination: filepath.Join(workbookRoot, "evidence", "acquired", "disk.raw")}
	dir := filepath.Join(workbookRoot, "hardware", "acquisitions")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(record)
	if err := os.WriteFile(filepath.Join(dir, record.OperationID+"-requested.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	overview, err := Build(root, "case-1")
	if err != nil {
		t.Fatal(err)
	}
	if overview.Acquisitions != 1 || overview.RecoveryRequired != 1 {
		t.Fatalf("overview=%+v", overview)
	}
}
