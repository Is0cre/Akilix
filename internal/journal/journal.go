package journal

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"time"
)

const Schema = "akilix.journal.v1"
const maxRecordSize = 1024 * 1024

var eventPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,63}$`)
var modulePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_-]{1,31}$`)

type Event struct {
	Schema       string         `json:"schema"`
	Timestamp    string         `json:"timestamp"`
	Event        string         `json:"event"`
	Module       string         `json:"module"`
	Payload      map[string]any `json:"payload"`
	ProvenanceID string         `json:"provenance_id"`
}

type Journal struct {
	path string
	mu   sync.Mutex
}

type Summary struct {
	DiscoveredHosts int `json:"discovered_hosts"`
	DiscoveredPorts int `json:"discovered_ports"`
	DroppedResults  int `json:"dropped_results"`
}

// Visit reads validated journal events under a shared file lock. The callback
// must not append to the same journal because doing so would deadlock.
func Visit(workbookRoot string, visit func(Event) error) error {
	if visit == nil {
		return fmt.Errorf("journal visitor is required")
	}
	path := filepath.Join(workbookRoot, "journal.jsonl")
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fd), path)
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("invalid workbook journal")
	}
	if err := syscall.Flock(fd, syscall.LOCK_SH); err != nil {
		return err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maxRecordSize+1)
	for scanner.Scan() {
		if len(scanner.Bytes()) > maxRecordSize {
			return fmt.Errorf("journal record exceeds size limit")
		}
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return fmt.Errorf("decode journal record: %w", err)
		}
		if event.Schema != Schema || !eventPattern.MatchString(event.Event) || !modulePattern.MatchString(event.Module) || event.Payload == nil || event.ProvenanceID == "" {
			return fmt.Errorf("invalid journal event")
		}
		if err := visit(event); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// Summarize reads canonical journal records without changing the workbook.
// Counts represent recorded observations, not a deduplicated asset inventory.
func Summarize(workbookRoot string) (Summary, error) {
	var result Summary
	err := Visit(workbookRoot, func(event Event) error {
		switch event.Event {
		case "HOST_DISCOVERED":
			result.DiscoveredHosts++
		case "PORT_FOUND":
			result.DiscoveredPorts++
		case "HOST_DROPPED_OUT_OF_SCOPE", "PORT_DROPPED_OUT_OF_SCOPE":
			result.DroppedResults++
		}
		return nil
	})
	return result, err
}

func Open(workbookRoot string) (*Journal, error) {
	info, err := os.Lstat(workbookRoot)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("invalid workbook journal root %q", workbookRoot)
	}
	path := filepath.Join(workbookRoot, "journal.jsonl")
	if info, err := os.Lstat(path); err == nil && (info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular()) {
		return nil, fmt.Errorf("invalid workbook journal path")
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return &Journal{path: path}, nil
}

func NewEvent(event, module string, payload map[string]any, now time.Time) (Event, error) {
	if !eventPattern.MatchString(event) || !modulePattern.MatchString(module) || payload == nil {
		return Event{}, fmt.Errorf("invalid journal event")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return Event{}, err
	}
	seed := append([]byte(now.UTC().Format(time.RFC3339Nano)+"\x00"+event+"\x00"+module), nonce[:]...)
	sum := sha256.Sum256(seed)
	return Event{Schema: Schema, Timestamp: now.UTC().Format("2006-01-02T15:04:05.000Z"), Event: event, Module: module, Payload: payload, ProvenanceID: "J-" + hex.EncodeToString(sum[:10])}, nil
}

func (j *Journal) Append(event Event) error {
	if event.Schema != Schema || event.Timestamp == "" || !eventPattern.MatchString(event.Event) || !modulePattern.MatchString(event.Module) || event.Payload == nil || event.ProvenanceID == "" {
		return fmt.Errorf("invalid journal event")
	}
	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if len(line) > maxRecordSize {
		return fmt.Errorf("journal event exceeds size limit")
	}
	line = append(line, '\n')
	j.mu.Lock()
	defer j.mu.Unlock()
	fd, err := syscall.Open(j.path, syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CREAT|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fd), j.path)
	defer file.Close()
	if err := syscall.Flock(fd, syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	info, err := file.Stat()
	if err != nil {
		return err
	}
	offset := info.Size()
	for len(line) != 0 {
		written, writeErr := file.Write(line)
		if writeErr != nil || written == 0 {
			_ = file.Truncate(offset)
			_ = file.Sync()
			if writeErr != nil {
				return writeErr
			}
			return fmt.Errorf("short journal write")
		}
		line = line[written:]
	}
	if err := file.Sync(); err != nil {
		_ = file.Truncate(offset)
		_ = file.Sync()
		return err
	}
	directory, err := os.Open(filepath.Dir(j.path))
	if err != nil {
		return err
	}
	err = directory.Sync()
	if closeErr := directory.Close(); err == nil {
		err = closeErr
	}
	return err
}

func (j *Journal) Path() string { return j.path }
