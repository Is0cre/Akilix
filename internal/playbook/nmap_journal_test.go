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

func TestIngestNmapJournalClassifiesHostsThroughScope(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "discovery.xml")
	data := `<?xml version="1.0"?><nmaprun><host><status state="up"/><address addr="192.0.2.10" addrtype="ipv4"/><address addr="aa:bb" addrtype="mac"/><hostnames><hostname name="router.test"/></hostnames></host><host><status state="up"/><address addr="198.51.100.8" addrtype="ipv4"/></host><host><status state="down"/><address addr="192.0.2.11" addrtype="ipv4"/></host></nmaprun>`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	log, err := journal.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	found, dropped, err := IngestNmapJournal(path, "inv-1", scope.Config{Includes: []string{"192.0.2.0/24"}}, log, func() time.Time { return time.Now().UTC() })
	if err != nil || found != 1 || dropped != 1 {
		t.Fatalf("found=%d dropped=%d err=%v", found, dropped, err)
	}
	file, err := os.Open(log.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var events []journal.Event
	for scanner.Scan() {
		var event journal.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	if len(events) != 2 || events[0].Event != "HOST_DISCOVERED" || events[0].Payload["hostname"] != "router.test" || events[1].Event != "HOST_DROPPED_OUT_OF_SCOPE" {
		t.Fatalf("events=%+v", events)
	}
}

func TestIngestNmapJournalRejectsSymlinkArtifact(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.xml")
	if err := os.WriteFile(target, []byte("<nmaprun/>"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.xml")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	log, _ := journal.Open(root)
	if _, _, err := IngestNmapJournal(link, "inv-1", scope.Config{}, log, time.Now); err == nil {
		t.Fatal("symlink artifact accepted")
	}
}

func TestIngestNmapJournalEmitsOnlyOpenAttributedPorts(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "ports.xml")
	data := `<?xml version="1.0"?><nmaprun><host><status state="up"/><address addr="2001:db8::10" addrtype="ipv6"/><ports><port protocol="tcp" portid="22"><state state="open"/><service name="ssh"/></port><port protocol="tcp" portid="23"><state state="closed"/></port></ports></host></nmaprun>`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	log, _ := journal.Open(root)
	if _, _, err := IngestNmapJournal(path, "inv-port", scope.Config{Includes: []string{"2001:db8::/64"}}, log, time.Now); err != nil {
		t.Fatal(err)
	}
	file, _ := os.Open(log.Path())
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var events []journal.Event
	for scanner.Scan() {
		var event journal.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	if len(events) != 2 || events[1].Event != "PORT_FOUND" || events[1].Payload["endpoint"] != "[2001:db8::10]:22" || events[1].Payload["service"] != "ssh" || events[1].Payload["invocation_id"] != "inv-port" {
		t.Fatalf("events=%+v", events)
	}
}
