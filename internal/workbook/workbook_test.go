package workbook

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateOpenList(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 8, 4, 20, 0, 0, 0, time.UTC)
	m, err := Create(root, "case-1", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.ID) != 36 || m.Status != "open" {
		t.Fatalf("bad metadata: %+v", m)
	}
	scopeData, err := os.ReadFile(filepath.Join(root, "case-1", "scope.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(scopeData) != "version: 1\ninclude:\n  []\nexclude:\n  []\n" {
		t.Fatalf("unexpected initial scope document: %q", scopeData)
	}
	if _, err := Open(root, "../escape"); err == nil {
		t.Fatal("path traversal accepted")
	}
	got, err := Open(root, "case-1")
	if err != nil || got.ID != m.ID {
		t.Fatalf("open: %+v %v", got, err)
	}
	list, err := List(root)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %+v %v", list, err)
	}
	if _, err := Create(root, "case-1", now); err == nil {
		t.Fatal("duplicate accepted")
	}
	closed, err := SetStatus(root, "case-1", "closed")
	if err != nil || closed.Status != "closed" {
		t.Fatalf("close: %+v %v", closed, err)
	}
	renamed, err := Rename(root, "case-1", "case-renamed")
	if err != nil || renamed.ID != m.ID || renamed.Name != "case-renamed" {
		t.Fatalf("rename: %+v %v", renamed, err)
	}
	if _, err := filepath.Abs(root); err != nil {
		t.Fatal(err)
	}
}

func TestOpenRejectsDirectoryMetadataMismatch(t *testing.T) {
	root := t.TempDir()
	if _, err := Create(root, "case-1", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "case-1", "workbook.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, bytes.Replace(b, []byte("name: case-1"), []byte("name: other"), -1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root, "case-1"); err == nil {
		t.Fatal("accepted metadata for another directory")
	}
}

func TestOpenRejectsSymlinkedWorkbookAndMetadata(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if _, err := Create(outside, "real", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "real"), filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root, "linked"); err == nil {
		t.Fatal("opened symlinked workbook directory")
	}
	if _, err := Create(root, "case-1", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(root, "case-1", "workbook.yaml")
	outsideMetadata := filepath.Join(outside, "workbook.yaml")
	b, err := os.ReadFile(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outsideMetadata, b, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(metadata); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideMetadata, metadata); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root, "case-1"); err == nil {
		t.Fatal("opened workbook with symlinked metadata")
	}
}

func TestListIgnoresInterruptedCreationDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".creating-interrupted"), 0700); err != nil {
		t.Fatal(err)
	}
	got, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("staging directory listed as workbook: %+v", got)
	}
}

func TestValidateLayoutRejectsMissingDirectory(t *testing.T) {
	root := t.TempDir()
	if _, err := Create(root, "case-1", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "case-1", "notes")); err != nil {
		t.Fatal(err)
	}
	if err := ValidateLayout(root, "case-1"); err == nil {
		t.Fatal("missing canonical directory accepted")
	}
}
