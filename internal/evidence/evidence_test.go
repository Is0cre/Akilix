package evidence

import (
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
