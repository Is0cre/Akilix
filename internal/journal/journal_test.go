package journal

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestConcurrentAppendProducesCompleteJSONLines(t *testing.T) {
	root := t.TempDir()
	journal, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	const count = 40
	var group sync.WaitGroup
	for i := 0; i < count; i++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			event, eventErr := NewEvent("PORT_FOUND", "RECON", map[string]any{"port": index}, time.Now())
			if eventErr != nil {
				t.Error(eventErr)
				return
			}
			if appendErr := journal.Append(event); appendErr != nil {
				t.Error(appendErr)
			}
		}(i)
	}
	group.Wait()
	file, err := os.Open(journal.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		if event.Schema != Schema || event.ProvenanceID == "" {
			t.Fatalf("event=%+v", event)
		}
		lines++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if lines != count {
		t.Fatalf("lines=%d want=%d", lines, count)
	}
	info, err := os.Stat(journal.Path())
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("mode=%v err=%v", info.Mode().Perm(), err)
	}
}

func TestJournalRejectsSymlinkAndInvalidTokens(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "journal.jsonl")); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root); err == nil {
		t.Fatal("symlink journal accepted")
	}
	if _, err := NewEvent("scope added", "CORE", map[string]any{}, time.Now()); err == nil {
		t.Fatal("invalid token accepted")
	}
}

func TestTimestampHasMillisecondUTCPrecision(t *testing.T) {
	event, err := NewEvent("SCOPE_TARGET_ADDED", "CORE", map[string]any{"value": "192.0.2.0/24"}, time.Date(2026, 9, 1, 1, 53, 12, 451999999, time.FixedZone("x", 3600)))
	if err != nil {
		t.Fatal(err)
	}
	if event.Timestamp != "2026-09-01T00:53:12.451Z" {
		t.Fatalf("timestamp=%q", event.Timestamp)
	}
}

func TestSummarizeCountsDiscoveryObservations(t *testing.T) {
	root := t.TempDir()
	log, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"HOST_DISCOVERED", "HOST_DISCOVERED", "PORT_FOUND", "HOST_DROPPED_OUT_OF_SCOPE", "INVOCATION_COMPLETED"} {
		event, err := NewEvent(name, "RECON", map[string]any{"test": true}, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		if err := log.Append(event); err != nil {
			t.Fatal(err)
		}
	}
	summary, err := Summarize(root)
	if err != nil || summary.DiscoveredHosts != 2 || summary.DiscoveredPorts != 1 || summary.DroppedResults != 1 {
		t.Fatalf("summary=%+v err=%v", summary, err)
	}
}

func TestSummarizeMissingJournalIsEmpty(t *testing.T) {
	summary, err := Summarize(t.TempDir())
	if err != nil || summary != (Summary{}) {
		t.Fatalf("summary=%+v err=%v", summary, err)
	}
}
