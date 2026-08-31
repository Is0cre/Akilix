package acquire

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const IdentificationSchema = "akilix.hardware-identification.v1"

type CommandResult struct {
	Tool     string          `json:"tool"`
	Args     []string        `json:"args"`
	ExitCode int             `json:"exit_code"`
	Status   string          `json:"status"`
	Output   json.RawMessage `json:"output,omitempty"`
	Error    string          `json:"error,omitempty"`
}

type Identification struct {
	Schema            string          `json:"schema"`
	ID                string          `json:"record_id,omitempty"`
	WorkbookID        string          `json:"workbook_id,omitempty"`
	RecordedAt        time.Time       `json:"recorded_at"`
	Device            Device          `json:"device"`
	InventoryRecordID string          `json:"inventory_record_id"`
	Passive           bool            `json:"passive"`
	Commands          []CommandResult `json:"commands"`
	RecordStatus      string          `json:"record_status,omitempty"`
}

func RecordIdentification(workbookRoot, workbookID string, identification Identification, now time.Time) (Identification, string, error) {
	if workbookID == "" || identification.Schema != IdentificationSchema || identification.InventoryRecordID == "" || !identification.Passive || len(identification.Commands) == 0 {
		return Identification{}, "", fmt.Errorf("invalid hardware identification provenance")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id, err := uuid7(now)
	if err != nil {
		return Identification{}, "", err
	}
	identification.ID = id
	identification.WorkbookID = workbookID
	identification.RecordStatus = "complete"
	data, err := json.MarshalIndent(identification, "", "  ")
	if err != nil {
		return Identification{}, "", err
	}
	data = append(data, '\n')
	hardwareDir := filepath.Join(workbookRoot, "hardware")
	identificationDir := filepath.Join(hardwareDir, "identifications")
	if err := ensureRealDirectory(hardwareDir); err != nil {
		return Identification{}, "", err
	}
	if err := ensureRealDirectory(identificationDir); err != nil {
		return Identification{}, "", err
	}
	path := filepath.Join(identificationDir, id+".json")
	if err := writeNewAtomic(path, data); err != nil {
		return Identification{}, "", err
	}
	return identification, path, nil
}

type ResultRunner interface {
	RunResult(context.Context, string, ...string) ([]byte, int, error)
}

type ExecResultRunner struct{}

func (ExecResultRunner) RunResult(ctx context.Context, name string, args ...string) ([]byte, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err == nil {
		return out, 0, nil
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return out, exit.ExitCode(), nil
	}
	if stderr.Len() != 0 {
		return nil, -1, fmt.Errorf("%s: %s", name, strings.TrimSpace(stderr.String()))
	}
	return nil, -1, err
}

func Identify(ctx context.Context, runner ResultRunner, device Device, inventoryRecordID string, now time.Time) (Identification, error) {
	if device.Path == "" || device.SystemDisk || !device.AcquisitionCandidate || device.Mounted || inventoryRecordID == "" {
		return Identification{}, fmt.Errorf("invalid identification candidate")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	identification := Identification{Schema: IdentificationSchema, RecordedAt: now.UTC(), Device: device, InventoryRecordID: inventoryRecordID, Passive: true, Commands: []CommandResult{}}
	commands := [][]string{{"smartctl", "--json=c", "--all", device.Path}}
	if device.Transport == "nvme" || strings.HasPrefix(device.KernelName, "nvme") {
		commands = [][]string{{"nvme", "id-ctrl", "--output-format=json", device.Path}, {"nvme", "smart-log", "--output-format=json", device.Path}}
	}
	for _, command := range commands {
		out, exitCode, runErr := runner.RunResult(ctx, command[0], command[1:]...)
		result := CommandResult{Tool: command[0], Args: append([]string(nil), command[1:]...), ExitCode: exitCode, Status: commandStatus(command[0], exitCode)}
		if runErr != nil {
			result.Status = "UNAVAILABLE"
			result.Error = runErr.Error()
			identification.Commands = append(identification.Commands, result)
			continue
		}
		if !json.Valid(out) {
			result.Status = "INVALID_OUTPUT"
			result.Error = "tool did not return valid JSON"
		} else {
			result.Output = append(json.RawMessage(nil), out...)
		}
		identification.Commands = append(identification.Commands, result)
	}
	return identification, nil
}

func commandStatus(tool string, exitCode int) string {
	if exitCode == 0 {
		return "OK"
	}
	if tool == "smartctl" && exitCode&8 != 0 {
		return "HEALTH_FAILED"
	}
	if tool == "smartctl" && exitCode&0xf8 != 0 {
		return "HEALTH_WARNING"
	}
	return "TOOL_ERROR"
}
