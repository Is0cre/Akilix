package acquire

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProtectionEventsAreImmutableAndCorrelated(t *testing.T) {
	root := t.TempDir()
	device := Device{Path: "/dev/sdb", Serial: "CASE-DISK", AcquisitionCandidate: true}
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	request, requestPath, err := RecordProtectionEvent(root, "workbook-id", "operation-id", "requested", device, ProtectionResult{}, nil, now)
	if err != nil || request.Phase != "REQUESTED" {
		t.Fatalf("request=%+v err=%v", request, err)
	}
	result := ProtectionResult{Device: device, KernelReadOnly: true}
	applied, appliedPath, err := RecordProtectionEvent(root, "workbook-id", "operation-id", "APPLIED", device, result, nil, now.Add(time.Second))
	if err != nil || !applied.KernelForcedRO || requestPath == appliedPath {
		t.Fatalf("applied=%+v paths=%q,%q err=%v", applied, requestPath, appliedPath, err)
	}
	for _, path := range []string{requestPath, appliedPath} {
		info, statErr := os.Stat(path)
		if statErr != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("path=%q mode=%v err=%v", path, info.Mode().Perm(), statErr)
		}
	}
}

func TestProtectionFailureLeavesRequestedEvent(t *testing.T) {
	root := t.TempDir()
	device := Device{Path: "/dev/sdb", AcquisitionCandidate: true}
	now := time.Now().UTC()
	_, requestPath, err := RecordProtectionEvent(root, "workbook-id", "operation-id", "REQUESTED", device, ProtectionResult{}, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	opErr := errors.New("permission denied")
	failed, _, err := RecordProtectionEvent(root, "workbook-id", "operation-id", "FAILED", device, ProtectionResult{}, opErr, now.Add(time.Second))
	if err != nil || failed.Error != opErr.Error() {
		t.Fatalf("failed=%+v err=%v", failed, err)
	}
	if _, err := os.Stat(requestPath); err != nil {
		t.Fatalf("requested event disappeared: %v", err)
	}
}

func TestProtectionEventRejectsInvalidStateAndSymlinkDirectory(t *testing.T) {
	device := Device{Path: "/dev/sdb"}
	if _, _, err := RecordProtectionEvent(t.TempDir(), "w", "o", "APPLIED", device, ProtectionResult{}, nil, time.Now()); err == nil {
		t.Fatal("unverified applied event accepted")
	}
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "hardware"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "hardware", "protections")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RecordProtectionEvent(root, "w", "o", "REQUESTED", device, ProtectionResult{}, nil, time.Now()); err == nil {
		t.Fatal("symlinked protection directory accepted")
	}
}
