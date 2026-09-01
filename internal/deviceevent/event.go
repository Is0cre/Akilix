package deviceevent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const Schema = "akilix.device-event.v1"

var deviceName = regexp.MustCompile(`^(sd[a-z]+|nvme[0-9]+n[0-9]+|mmcblk[0-9]+)$`)

type Event struct {
	Schema           string    `json:"schema"`
	Timestamp        time.Time `json:"timestamp"`
	Action           string    `json:"action"`
	Device           string    `json:"device"`
	Serial           string    `json:"serial,omitempty"`
	Vendor           string    `json:"vendor,omitempty"`
	Model            string    `json:"model,omitempty"`
	Bus              string    `json:"bus,omitempty"`
	KernelForcedRO   bool      `json:"kernel_forced_ro"`
	OperatorDecision string    `json:"operator_decision"`
}

func New(action, device string, properties map[string]string, kernelRO bool, now time.Time) (Event, error) {
	if action != "add" && action != "remove" {
		return Event{}, fmt.Errorf("unsupported device action %q", action)
	}
	base := filepath.Base(device)
	if device != filepath.Join("/dev", base) || !deviceName.MatchString(base) {
		return Event{}, fmt.Errorf("invalid whole-disk device %q", device)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return Event{Schema: Schema, Timestamp: now.UTC(), Action: action, Device: device,
		Serial: clean(properties["ID_SERIAL_SHORT"]), Vendor: clean(properties["ID_VENDOR"]),
		Model: clean(properties["ID_MODEL"]), Bus: clean(properties["ID_BUS"]),
		KernelForcedRO: kernelRO, OperatorDecision: "PENDING"}, nil
}

func Append(queue string, event Event) (string, error) {
	if event.Schema != Schema || event.OperatorDecision != "PENDING" {
		return "", fmt.Errorf("invalid device event")
	}
	if err := os.MkdirAll(queue, 0770); err != nil {
		return "", err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("%d-%s-%s.json", event.Timestamp.UnixNano(), event.Action, filepath.Base(event.Device))
	tmp, err := os.CreateTemp(queue, ".device-event-")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err = tmp.Chmod(0640); err == nil {
		_, err = tmp.Write(append(data, '\n'))
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	path := filepath.Join(queue, name)
	if err == nil {
		err = os.Rename(tmpPath, path)
	}
	return path, err
}

func clean(value string) string { return strings.TrimSpace(strings.ReplaceAll(value, "_", " ")) }
