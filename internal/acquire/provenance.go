package acquire

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const ProvenanceSchema = "akilix.hardware-provenance.v1"

type ProvenanceRecord struct {
	Schema       string     `json:"schema"`
	ID           string     `json:"record_id"`
	WorkbookID   string     `json:"workbook_id"`
	Event        string     `json:"event"`
	RecordedAt   time.Time  `json:"recorded_at"`
	Inspection   Inspection `json:"inspection"`
	RecordStatus string     `json:"record_status"`
}

// RecordInspection durably records an already completed passive inspection.
// It creates a new immutable file and never overwrites an existing record.
func RecordInspection(workbookRoot, workbookID string, inspection Inspection, now time.Time) (ProvenanceRecord, string, error) {
	if workbookID == "" || inspection.Schema != InspectionSchema || !inspection.Passive {
		return ProvenanceRecord{}, "", fmt.Errorf("invalid hardware provenance input")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id, err := uuid7(now)
	if err != nil {
		return ProvenanceRecord{}, "", err
	}
	record := ProvenanceRecord{Schema: ProvenanceSchema, ID: id, WorkbookID: workbookID, Event: "INVENTORY_RECORDED", RecordedAt: now.UTC(), Inspection: inspection, RecordStatus: "complete"}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return ProvenanceRecord{}, "", err
	}
	data = append(data, '\n')
	hardwareDir := filepath.Join(workbookRoot, "hardware")
	inspectionDir := filepath.Join(hardwareDir, "inspections")
	if err := ensureRealDirectory(hardwareDir); err != nil {
		return ProvenanceRecord{}, "", err
	}
	if err := ensureRealDirectory(inspectionDir); err != nil {
		return ProvenanceRecord{}, "", err
	}
	path := filepath.Join(inspectionDir, id+".json")
	if err := writeNewAtomic(path, data); err != nil {
		return ProvenanceRecord{}, "", err
	}
	return record, path, nil
}

func ensureRealDirectory(path string) error {
	err := os.Mkdir(path, 0700)
	if err != nil && !os.IsExist(err) {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("invalid hardware provenance directory %q", path)
	}
	return nil
}

func writeNewAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".record-")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Link(tmpPath, path); err != nil {
		return fmt.Errorf("create immutable hardware record: %w", err)
	}
	if err := os.Remove(tmpPath); err != nil {
		_ = os.Remove(path)
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = d.Sync()
	if closeErr := d.Close(); err == nil {
		err = closeErr
	}
	return err
}

func uuid7(t time.Time) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[6:]); err != nil {
		return "", err
	}
	ms := uint64(t.UnixMilli())
	b[0], b[1], b[2], b[3], b[4], b[5] = byte(ms>>40), byte(ms>>32), byte(ms>>24), byte(ms>>16), byte(ms>>8), byte(ms)
	b[6] = (b[6] & 0x0f) | 0x70
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16])), nil
}
