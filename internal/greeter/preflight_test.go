package greeter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixture(t *testing.T, root, path, value string) {
	t.Helper()
	full := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(value), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestInspectUsesReadOnlyLocalStateAndRedactsWorkbooks(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "sys/fs/selinux/enforce", "1\n")
	writeFixture(t, root, "sys/class/block/sdb/removable", "1\n")
	writeFixture(t, root, "sys/class/block/sdb/ro", "1\n")
	writeFixture(t, root, "usr/bin/podman", "")
	writeFixture(t, root, "etc/subuid", "akilix:100000:65536\n")
	snapshot := Inspect(root)
	if snapshot.SELinux != "ENFORCING" || snapshot.Media != "1 peripheral media · READ-ONLY" || snapshot.MediaWarning || snapshot.Containers != "Podman · rootless configured" {
		t.Fatalf("snapshot=%+v", snapshot)
	}
	if snapshot.WorkbookState != "Available after authentication" {
		t.Fatalf("pre-auth workbook data leaked: %+v", snapshot)
	}
}

func TestInspectWarnsForWritablePeripheralAndIgnoresPartitions(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "sys/class/block/sdc/removable", "1")
	writeFixture(t, root, "sys/class/block/sdc/ro", "0")
	writeFixture(t, root, "sys/class/block/sdc1/removable", "1")
	writeFixture(t, root, "sys/class/block/sdc1/ro", "0")
	writeFixture(t, root, "sys/class/block/sdc1/partition", "1")
	snapshot := Inspect(root)
	if !snapshot.MediaWarning || snapshot.Media != "1 peripheral media · 1 NOT READ-ONLY" {
		t.Fatalf("snapshot=%+v", snapshot)
	}
}

func TestRenderCreatesDualColumnCardAndSymmetricHints(t *testing.T) {
	view := Render(Preflight{SELinux: "ENFORCING", Media: "No peripheral media", Containers: "Podman · rootless configured", WorkbookState: "Available after authentication"}, false)
	for _, marker := range []string{"AKILIX", "SYSTEM PRE-FLIGHT AUDIT", "SELinux: ENFORCING", "Available after authentication", "ESC Reset", "F3 Session"} {
		if !strings.Contains(view, marker) {
			t.Fatalf("view lacks %q:\n%s", marker, view)
		}
	}
	if strings.Contains(view, "client-") || strings.Contains(view, "/dev/") {
		t.Fatalf("pre-auth view leaks sensitive names: %s", view)
	}
	if strings.Contains(view, "\x1b[") {
		t.Fatalf("no-color view contains terminal control sequences: %q", view)
	}
}
