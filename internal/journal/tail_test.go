package journal

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTailReturnsOnlyCompleteNewLinesAndHandlesTruncation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	if err := os.WriteFile(path, []byte("one\ntwo"), 0600); err != nil {
		t.Fatal(err)
	}
	tail := NewTail(path)
	lines, err := tail.Poll()
	if err != nil || !reflect.DeepEqual(lines, []string{"one"}) {
		t.Fatalf("lines=%v err=%v", lines, err)
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("\nthree\n"); err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	lines, err = tail.Poll()
	if err != nil || !reflect.DeepEqual(lines, []string{"two", "three"}) {
		t.Fatalf("lines=%v err=%v", lines, err)
	}
	if err := os.WriteFile(path, []byte("new\n"), 0600); err != nil {
		t.Fatal(err)
	}
	lines, err = tail.Poll()
	if err != nil || !reflect.DeepEqual(lines, []string{"new"}) {
		t.Fatalf("truncated lines=%v err=%v", lines, err)
	}
}

func TestTailRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("x\n"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "journal.jsonl")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := NewTail(link).Poll(); err == nil {
		t.Fatal("symlink tail accepted")
	}
}
