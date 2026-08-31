package container

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOriginalEvidenceMountIsFixedAndReadOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "evidence", "original"), 0700); err != nil {
		t.Fatal(err)
	}
	mount, err := OriginalEvidenceMount(root)
	if err != nil {
		t.Fatal(err)
	}
	if !mount.ReadOnly || !mount.OriginalEvidence || mount.Destination != "/workbook/evidence/original" {
		t.Fatalf("unsafe mount: %+v", mount)
	}
}

func TestOriginalEvidenceMountRejectsSymlink(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "evidence"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "evidence", "original")); err != nil {
		t.Fatal(err)
	}
	if _, err := OriginalEvidenceMount(root); err == nil {
		t.Fatal("symlinked original evidence path accepted")
	}
}
