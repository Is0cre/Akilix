package activity

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const Schema = "pensuse.activity.v1"

type Event struct {
	Schema       string    `json:"schema"`
	InvocationID string    `json:"invocation_id"`
	WorkbookID   string    `json:"workbook_id"`
	Timestamp    time.Time `json:"timestamp"`
	Phase        string    `json:"phase"`
	Executor     string    `json:"executor"`
	Tool         string    `json:"tool"`
	ExitCode     *int      `json:"exit_code,omitempty"`
}

func Append(root string, event Event) error {
	if event.Schema != Schema || event.InvocationID == "" || event.WorkbookID == "" || event.Timestamp.IsZero() || event.Tool == "" || event.Phase != "STARTED" && event.Phase != "COMPLETED" && event.Phase != "FAILED" {
		return fmt.Errorf("invalid activity event")
	}
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	dir := filepath.Join(root, ".pensuse")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "activity.jsonl"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	if _, err := f.Write(b); err != nil {
		return err
	}
	return f.Sync()
}

func List(root string) ([]Event, error) {
	f, err := os.Open(filepath.Join(root, ".pensuse", "activity.jsonl"))
	if os.IsNotExist(err) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return nil, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var events []Event
	for len(b) > 0 {
		i := 0
		for i < len(b) && b[i] != '\n' {
			i++
		}
		if i == len(b) {
			return nil, fmt.Errorf("partial activity record")
		}
		var event Event
		if err := json.Unmarshal(b[:i], &event); err != nil {
			return nil, err
		}
		events = append(events, event)
		b = b[i+1:]
	}
	return events, nil
}
