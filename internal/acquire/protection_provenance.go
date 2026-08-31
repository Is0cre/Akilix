package acquire

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const ProtectionSchema = "akilix.hardware-protection.v1"

type ProtectionEvent struct {
	Schema          string    `json:"schema"`
	ID              string    `json:"record_id"`
	OperationID     string    `json:"operation_id"`
	WorkbookID      string    `json:"workbook_id"`
	Phase           string    `json:"phase"`
	RecordedAt      time.Time `json:"recorded_at"`
	Device          Device    `json:"device"`
	KernelForcedRO  bool      `json:"kernel_forced_ro"`
	AlreadyReadOnly bool      `json:"already_read_only,omitempty"`
	Error           string    `json:"error,omitempty"`
	RecordStatus    string    `json:"record_status"`
}

func NewOperationID(now time.Time) (string, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return uuid7(now)
}

// RecordProtectionEvent creates one immutable event in a protection operation.
// REQUESTED must be recorded before attempting the kernel state change.
func RecordProtectionEvent(workbookRoot, workbookID, operationID, phase string, device Device, result ProtectionResult, operationErr error, now time.Time) (ProtectionEvent, string, error) {
	if workbookID == "" || operationID == "" || device.Path == "" {
		return ProtectionEvent{}, "", fmt.Errorf("invalid hardware protection provenance input")
	}
	phase = strings.ToUpper(phase)
	if phase != "REQUESTED" && phase != "APPLIED" && phase != "FAILED" {
		return ProtectionEvent{}, "", fmt.Errorf("invalid hardware protection phase %q", phase)
	}
	if phase == "FAILED" && operationErr == nil {
		return ProtectionEvent{}, "", fmt.Errorf("failed protection event requires an error")
	}
	if phase == "APPLIED" && (!result.KernelReadOnly || operationErr != nil) {
		return ProtectionEvent{}, "", fmt.Errorf("applied protection event requires verified read-only state")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id, err := uuid7(now)
	if err != nil {
		return ProtectionEvent{}, "", err
	}
	record := ProtectionEvent{
		Schema:          ProtectionSchema,
		ID:              id,
		OperationID:     operationID,
		WorkbookID:      workbookID,
		Phase:           phase,
		RecordedAt:      now.UTC(),
		Device:          device,
		KernelForcedRO:  result.KernelReadOnly,
		AlreadyReadOnly: result.AlreadyReadOnly,
		RecordStatus:    "complete",
	}
	if operationErr != nil {
		record.Error = operationErr.Error()
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return ProtectionEvent{}, "", err
	}
	data = append(data, '\n')
	hardwareDir := filepath.Join(workbookRoot, "hardware")
	protectionDir := filepath.Join(hardwareDir, "protections")
	if err := ensureRealDirectory(hardwareDir); err != nil {
		return ProtectionEvent{}, "", err
	}
	if err := ensureRealDirectory(protectionDir); err != nil {
		return ProtectionEvent{}, "", err
	}
	path := filepath.Join(protectionDir, id+".json")
	if err := writeNewAtomic(path, data); err != nil {
		return ProtectionEvent{}, "", err
	}
	return record, path, nil
}
