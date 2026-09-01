package playbook

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/scope"
)

func TestIngestNaabuJournalClassifiesPortsThroughScope(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "ports.jsonl")
	data := "{\"ip\":\"192.0.2.10\",\"port\":53,\"protocol\":\"tcp\"}\n{\"ip\":\"198.51.100.8\",\"port\":443}\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	log, err := journal.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	found, dropped, err := IngestNaabuJournal(path, "inv-1", scope.Config{Includes: []string{"192.0.2.0/24"}}, log, func() time.Time { return time.Now().UTC() })
	if err != nil || found != 1 || dropped != 1 {
		t.Fatalf("found=%d dropped=%d err=%v", found, dropped, err)
	}
	file, err := os.Open(log.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var names []string
	for scanner.Scan() {
		var event journal.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		names = append(names, event.Event)
	}
	if len(names) != 2 || names[0] != "PORT_FOUND" || names[1] != "PORT_DROPPED_OUT_OF_SCOPE" {
		t.Fatalf("events=%v", names)
	}
}
