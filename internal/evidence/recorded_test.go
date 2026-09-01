package evidence

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Is0cre/Akilix/internal/journal"
)

func TestRecordedEvidenceLifecycleJournalsRequestImportAndVerification(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workbook")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(source, []byte("evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	record, err := ImportRecorded(root, "wb-1", source, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if ok, _, err := VerifyRecorded(root, "wb-1", record.ID, time.Now()); err != nil || !ok {
		t.Fatalf("ok=%t err=%v", ok, err)
	}
	f, err := os.Open(filepath.Join(root, "journal.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var event journal.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		names = append(names, event.Event)
	}
	want := []string{"EVIDENCE_IMPORT_REQUESTED", "EVIDENCE_IMPORTED", "EVIDENCE_VERIFIED"}
	if len(names) != len(want) {
		t.Fatalf("events=%v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("events=%v", names)
		}
	}
}

func TestRecordedImportFailureIsVisible(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workbook")
	_ = os.MkdirAll(root, 0700)
	if _, err := ImportRecorded(root, "wb-1", filepath.Join(t.TempDir(), "missing.bin"), time.Now()); err == nil {
		t.Fatal("missing source imported")
	}
	data, err := os.ReadFile(filepath.Join(root, "journal.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("EVIDENCE_IMPORT_FAILED")) {
		t.Fatalf("journal=%s", data)
	}
}
