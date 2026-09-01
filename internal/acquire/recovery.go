package acquire

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ImageOperationStatus struct {
	OperationID       string `json:"operation_id"`
	Source            string `json:"source"`
	Destination       string `json:"destination"`
	State             string `json:"state"`
	SizeBytes         int64  `json:"size_bytes,omitempty"`
	SHA256            string `json:"sha256,omitempty"`
	Verification      string `json:"verification,omitempty"`
	VerificationCount int    `json:"verification_count"`
	RecoveryRequired  bool   `json:"recovery_required"`
}

// ImageStatus reconstructs acquisition state solely from immutable records.
// It never opens source devices or changes/removes partial files.
func ImageStatus(workbookRoot, operationID string) ([]ImageOperationStatus, error) {
	if operationID != "" && !validOperationID(operationID) {
		return nil, fmt.Errorf("invalid acquisition operation ID")
	}
	dir := filepath.Join(workbookRoot, "hardware", "acquisitions")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []ImageOperationStatus{}, nil
	}
	if err != nil {
		return nil, err
	}
	states := map[string]*ImageOperationStatus{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := readRegularNoFollow(path)
		if err != nil {
			return nil, err
		}
		var record ImageRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("decode acquisition record %q: %w", entry.Name(), err)
		}
		if record.Schema != ImageSchema || !validOperationID(record.OperationID) || (operationID != "" && record.OperationID != operationID) {
			if operationID != "" {
				continue
			}
			return nil, fmt.Errorf("invalid acquisition record %q", entry.Name())
		}
		state := states[record.OperationID]
		if state == nil {
			state = &ImageOperationStatus{OperationID: record.OperationID, Source: record.Source, Destination: record.Destination, State: "REQUESTED", RecoveryRequired: true}
			states[record.OperationID] = state
		}
		switch record.Phase {
		case "COMPLETED":
			state.State, state.SizeBytes, state.SHA256, state.RecoveryRequired = "COMPLETED", record.SizeBytes, record.SHA256, false
		case "FAILED":
			if state.State != "COMPLETED" {
				state.State, state.RecoveryRequired = "FAILED", true
			}
		case "VERIFIED":
			state.Verification, state.VerificationCount = record.Verification, state.VerificationCount+1
		case "REQUESTED":
		default:
			return nil, fmt.Errorf("invalid acquisition phase %q", record.Phase)
		}
	}
	out := make([]ImageOperationStatus, 0, len(states))
	for _, state := range states {
		out = append(out, *state)
	}
	sort.Slice(out, func(i, j int) bool { return strings.Compare(out[i].OperationID, out[j].OperationID) < 0 })
	if operationID != "" && len(out) == 0 {
		return nil, fmt.Errorf("acquisition %s not found", operationID)
	}
	return out, nil
}
