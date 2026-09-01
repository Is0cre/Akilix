package acquire

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestImageStatusDistinguishesCompletedFailedAndRequested(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "evidence", "acquired"), 0700)
	completed, _, err := imageFromReader(context.Background(), root, "wb", "/dev/a", "a.raw", bytes.NewReader([]byte("abc")), 3, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	_, _, _ = imageFromReader(context.Background(), root, "wb", "/dev/b", "b.raw", bytes.NewReader([]byte("x")), 2, time.Now().Add(time.Millisecond))
	requested := ImageRecord{Schema: ImageSchema, OperationID: "018f1f1e-7b8a-7000-8000-000000000001", WorkbookID: "wb", Phase: "REQUESTED", RecordedAt: time.Now(), Source: "/dev/c", Destination: filepath.Join(root, "evidence", "acquired", "c.raw")}
	if err := recordImage(root, requested); err != nil {
		t.Fatal(err)
	}
	states, err := ImageStatus(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 3 {
		t.Fatalf("states=%+v", states)
	}
	byID := map[string]ImageOperationStatus{}
	for _, state := range states {
		byID[state.OperationID] = state
	}
	if byID[completed.OperationID].State != "COMPLETED" || byID[completed.OperationID].RecoveryRequired {
		t.Fatalf("completed=%+v", byID[completed.OperationID])
	}
	if byID[requested.OperationID].State != "REQUESTED" || !byID[requested.OperationID].RecoveryRequired {
		t.Fatalf("requested=%+v", byID[requested.OperationID])
	}
}

func TestImageStatusRejectsSymlinkRecord(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "hardware", "acquisitions")
	_ = os.MkdirAll(dir, 0700)
	outside := filepath.Join(t.TempDir(), "x.json")
	_ = os.WriteFile(outside, []byte("{}"), 0600)
	_ = os.Symlink(outside, filepath.Join(dir, "x.json"))
	if _, err := ImageStatus(root, ""); err == nil {
		t.Fatal("symlink record accepted")
	}
}
