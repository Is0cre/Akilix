package acquire

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestImageFromReaderCompletesWithHash(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "evidence", "acquired"), 0700); err != nil {
		t.Fatal(err)
	}
	payload := bytes.Repeat([]byte("akilix"), 1000)
	record, path, err := imageFromReader(context.Background(), root, "wb-1", "/dev/test", "disk.raw", bytes.NewReader(payload), int64(len(payload)), time.Date(2026, 9, 1, 2, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256(payload)
	if record.Phase != "COMPLETED" || record.SizeBytes != int64(len(payload)) || record.SHA256 != hex.EncodeToString(want[:]) {
		t.Fatalf("record=%+v", record)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("image content mismatch")
	}
	if _, err := os.Stat(filepath.Join(root, "hardware", "acquisitions", record.OperationID+"-requested.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "hardware", "acquisitions", record.OperationID+"-completed.json")); err != nil {
		t.Fatal(err)
	}
}

func TestImageFromReaderCancellationLeavesNoImage(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "evidence", "acquired"), 0700)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	record, _, err := imageFromReader(ctx, root, "wb-1", "/dev/test", "disk.raw", bytes.NewReader([]byte("data")), 4, time.Now())
	if err == nil || record.Phase != "FAILED" {
		t.Fatalf("record=%+v err=%v", record, err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "evidence", "acquired", "disk.raw")); !os.IsNotExist(statErr) {
		t.Fatalf("destination exists: %v", statErr)
	}
}

func TestImageRejectsTraversalAndOverwrite(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "evidence", "acquired"), 0700)
	if _, _, err := imageFromReader(context.Background(), root, "wb", "/dev/test", "../disk.raw", bytes.NewReader(nil), 0, time.Now()); err == nil {
		t.Fatal("traversal accepted")
	}
}

func TestImageFromReaderRefusesShortSource(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "evidence", "acquired"), 0700)
	record, _, err := imageFromReader(context.Background(), root, "wb", "/dev/test", "disk.raw", bytes.NewReader([]byte("short")), 10, time.Now())
	if err == nil || record.Phase != "FAILED" {
		t.Fatalf("record=%+v err=%v", record, err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "evidence", "acquired", "disk.raw")); !os.IsNotExist(statErr) {
		t.Fatalf("partial image published: %v", statErr)
	}
}

func TestVerifyImageDetectsMatchAndMismatch(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "evidence", "acquired"), 0700)
	payload := []byte("forensic-image")
	completed, path, err := imageFromReader(context.Background(), root, "wb", "/dev/test", "disk.raw", bytes.NewReader(payload), int64(len(payload)), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	verified, match, err := VerifyImage(context.Background(), root, "wb", completed.OperationID, time.Now())
	if err != nil || !match || verified.Verification != "match" {
		t.Fatalf("verified=%+v match=%t err=%v", verified, match, err)
	}
	if err := os.WriteFile(path, []byte("tampered-image"), 0600); err != nil {
		t.Fatal(err)
	}
	verified, match, err = VerifyImage(context.Background(), root, "wb", completed.OperationID, time.Now().Add(time.Second))
	if err != nil || match || verified.Verification != "mismatch" {
		t.Fatalf("verified=%+v match=%t err=%v", verified, match, err)
	}
}

func TestVerifyImageReportsIncomplete(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "hardware", "acquisitions"), 0700)
	_, _, err := VerifyImage(context.Background(), root, "wb", "018f1f1e-7b8a-7000-8000-000000000001", time.Now())
	if err == nil {
		t.Fatal("incomplete acquisition accepted")
	}
}
