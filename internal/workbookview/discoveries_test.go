package workbookview

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/workbook"
)

func TestDiscoveriesGroupsCanonicalJournalObservations(t *testing.T) {
	root := t.TempDir()
	if _, err := workbook.Create(root, "case-1", time.Now()); err != nil {
		t.Fatal(err)
	}
	log, _ := journal.Open(filepath.Join(root, "case-1"))
	for index, payload := range []map[string]any{
		{"address": "192.0.2.10", "hostname": "first.test", "invocation_id": "inv-1"},
		{"address": "192.0.2.10", "hostname": "latest.test", "invocation_id": "inv-2"},
		{"endpoint": "192.0.2.10:443", "invocation_id": "inv-3"},
	} {
		name := "HOST_DISCOVERED"
		if index == 2 {
			name = "PORT_FOUND"
		}
		event, _ := journal.NewEvent(name, "RECON", payload, time.Date(2026, 9, 1, 1, index, 0, 0, time.UTC))
		if err := log.Append(event); err != nil {
			t.Fatal(err)
		}
	}
	items, err := Discoveries(root, "case-1")
	if err != nil || len(items) != 2 || items[0].Kind != "host" || items[0].Occurrences != 2 || items[0].Hostname != "latest.test" || items[0].LastInvocationID != "inv-2" || items[1].Kind != "port" {
		t.Fatalf("items=%+v err=%v", items, err)
	}
}
