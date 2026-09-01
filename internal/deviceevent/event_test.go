package deviceevent

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestWholeDiskEventIsQueuedAsPending(t *testing.T) {
	event, err := New("add", "/dev/sdb", map[string]string{"ID_SERIAL_SHORT": "ABC", "ID_VENDOR": "ACME_CORP", "ID_MODEL": "Evidence_Disk", "ID_BUS": "usb"}, true, time.Date(2026, 9, 1, 2, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	path, err := Append(t.TempDir(), event)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{`"device":"/dev/sdb"`, `"kernel_forced_ro":true`, `"operator_decision":"PENDING"`, `"vendor":"ACME CORP"`} {
		if !strings.Contains(string(data), marker) {
			t.Fatalf("event lacks %q: %s", marker, data)
		}
	}
}

func TestEventRejectsPartitionsAndArbitraryPaths(t *testing.T) {
	for _, path := range []string{"/dev/sdb1", "/tmp/sdb", "/dev/../dev/sdb", "/dev/mapper/test"} {
		if _, err := New("add", path, nil, true, time.Now()); err == nil {
			t.Fatalf("accepted %q", path)
		}
	}
}
