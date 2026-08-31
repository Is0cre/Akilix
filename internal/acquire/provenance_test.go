package acquire

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecordInspectionCreatesCanonicalImmutableRecord(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 8, 31, 14, 0, 0, 0, time.UTC)
	inspection := Inspection{Schema: InspectionSchema, GeneratedAt: now.Add(-time.Second), Passive: true, Source: "lsblk-json", Devices: []Device{}}
	record, path, err := RecordInspection(root, "workbook-id", inspection, now)
	if err != nil {
		t.Fatal(err)
	}
	if record.Schema != ProvenanceSchema || record.Event != "INVENTORY_RECORDED" || record.RecordStatus != "complete" || filepath.Dir(path) != filepath.Join(root, "hardware", "inspections") {
		t.Fatalf("unexpected record: %+v path=%s", record, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ProvenanceRecord
	if err := json.Unmarshal(data, &decoded); err != nil || decoded.ID != record.ID || decoded.WorkbookID != "workbook-id" {
		t.Fatalf("stored record invalid: %+v %v", decoded, err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("record mode: %v %v", info, err)
	}
}

func TestRecordInspectionRejectsSymlinkedHardwareDirectory(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "hardware")); err != nil {
		t.Fatal(err)
	}
	inspection := Inspection{Schema: InspectionSchema, Passive: true, Source: "lsblk-json", Devices: []Device{}}
	if _, _, err := RecordInspection(root, "workbook-id", inspection, time.Now()); err == nil {
		t.Fatal("symlinked hardware directory accepted")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("wrote outside workbook: %v %v", entries, err)
	}
}

func TestRecordInspectionRejectsNonPassiveInput(t *testing.T) {
	if _, _, err := RecordInspection(t.TempDir(), "workbook-id", Inspection{Schema: InspectionSchema}, time.Now()); err == nil {
		t.Fatal("non-passive inspection accepted")
	}
}
