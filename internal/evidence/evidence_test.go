package evidence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestImportVerifyAndNoOverwrite(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.bin")
	if err := os.WriteFile(source, []byte("evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	workbook := filepath.Join(root, "workbook")
	if err := os.MkdirAll(workbook, 0700); err != nil {
		t.Fatal(err)
	}
	r, err := Import(workbook, source, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != "complete" || r.SHA256 == "" {
		t.Fatalf("bad record: %+v", r)
	}
	ok, verified, err := Verify(workbook, r.ID)
	if err != nil || !ok {
		t.Fatalf("verify: %v %v", ok, err)
	}
	if verified.Verification != "match" {
		t.Fatalf("verification status: %+v", verified)
	}
	if _, err := Import(workbook, source, time.Unix(1, 0)); err == nil {
		t.Fatal("duplicate original accepted")
	}
	if _, err := os.Stat(filepath.Join(workbook, "evidence", "manifests", r.ID+".json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workbook, "evidence", "original", r.Filename), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	ok, verified, err = Verify(workbook, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("tampered evidence verified")
	}
	if verified.Verification != "mismatch" {
		t.Fatalf("mismatch status: %+v", verified)
	}
}

func TestImportRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "real")
	link := filepath.Join(root, "link")
	if err := os.WriteFile(source, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, link); err != nil {
		t.Fatal(err)
	}
	if _, err := Import(filepath.Join(root, "workbook"), link, time.Now()); err == nil {
		t.Fatal("symlink accepted")
	}
}

func TestVerifyRejectsManifestPathTraversal(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.bin")
	if err := os.WriteFile(source, []byte("evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	workbook := filepath.Join(root, "workbook")
	if _, err := Import(workbook, source, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	list, err := List(workbook)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %+v %v", list, err)
	}
	record := list[0]
	record.Filename = "../../source.bin"
	b, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(workbook, "evidence", "manifests", record.ID+".json")
	if err := os.WriteFile(manifest, b, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Verify(workbook, record.ID); err == nil {
		t.Fatal("accepted path traversal in evidence manifest")
	}
}

func TestListRejectsManifestIDMismatch(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.bin")
	if err := os.WriteFile(source, []byte("evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	workbook := filepath.Join(root, "workbook")
	record, err := Import(workbook, source, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	wrong := "77777777-7777-7777-8777-777777777777"
	oldPath := filepath.Join(workbook, "evidence", "manifests", record.ID+".json")
	newPath := filepath.Join(workbook, "evidence", "manifests", wrong+".json")
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	if _, err := List(workbook); err == nil {
		t.Fatal("accepted manifest filename mismatch")
	}
}

func TestVerifyAllReportsMismatch(t *testing.T) {
	root := t.TempDir()
	workbook := filepath.Join(root, "workbook")
	source := filepath.Join(root, "source.bin")
	if err := os.WriteFile(source, []byte("evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Import(workbook, source, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	records, err := List(workbook)
	if err != nil || len(records) != 1 {
		t.Fatalf("list: %+v %v", records, err)
	}
	if err := os.WriteFile(filepath.Join(workbook, "evidence", "original", records[0].Filename), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	verified, allMatch, err := VerifyAll(workbook)
	if err != nil || allMatch || len(verified) != 1 || verified[0].Verification != "mismatch" {
		t.Fatalf("verify all: %+v %v %v", verified, allMatch, err)
	}
}

func TestCheckAllIsReadOnly(t *testing.T) {
	root := t.TempDir()
	workbook := filepath.Join(root, "workbook")
	source := filepath.Join(root, "source.bin")
	if err := os.WriteFile(source, []byte("evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	record, err := Import(workbook, source, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(workbook, "evidence", "manifests", record.ID+".json")
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	checked, allMatch, err := CheckAll(workbook)
	if err != nil || !allMatch || len(checked) != 1 || checked[0].Verification != "match" {
		t.Fatalf("check all: %+v %v %v", checked, allMatch, err)
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("read-only evidence check modified manifest")
	}
}
