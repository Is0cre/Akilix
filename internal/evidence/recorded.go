package evidence

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Is0cre/Akilix/internal/journal"
)

func ImportRecorded(workbookRoot, workbookID, source string, now time.Time) (Record, error) {
	if workbookID == "" {
		return Record{}, fmt.Errorf("workbook identity is required")
	}
	log, err := journal.Open(workbookRoot)
	if err != nil {
		return Record{}, err
	}
	requested, err := journal.NewEvent("EVIDENCE_IMPORT_REQUESTED", "EVIDENCE", map[string]any{"workbook_id": workbookID, "source_name": filepath.Base(source), "operator_action": "EXPLICIT_CONFIRMED"}, now)
	if err != nil {
		return Record{}, err
	}
	if err := log.Append(requested); err != nil {
		return Record{}, err
	}
	record, importErr := Import(workbookRoot, source, now)
	if importErr != nil {
		failed, eventErr := journal.NewEvent("EVIDENCE_IMPORT_FAILED", "EVIDENCE", map[string]any{"workbook_id": workbookID, "source_name": filepath.Base(source), "request_provenance_id": requested.ProvenanceID, "error": importErr.Error()}, time.Now().UTC())
		if eventErr == nil {
			eventErr = log.Append(failed)
		}
		if eventErr != nil {
			return Record{}, fmt.Errorf("%v; journal failure: %w", importErr, eventErr)
		}
		return Record{}, importErr
	}
	completed, err := journal.NewEvent("EVIDENCE_IMPORTED", "EVIDENCE", map[string]any{"workbook_id": workbookID, "evidence_id": record.ID, "filename": record.Filename, "size_bytes": record.Size, "sha256": record.SHA256, "request_provenance_id": requested.ProvenanceID}, time.Now().UTC())
	if err != nil {
		return record, err
	}
	if err := log.Append(completed); err != nil {
		return record, fmt.Errorf("evidence imported but journal completion failed: %w", err)
	}
	return record, nil
}

func VerifyRecorded(workbookRoot, workbookID, evidenceID string, now time.Time) (bool, Record, error) {
	if workbookID == "" {
		return false, Record{}, fmt.Errorf("workbook identity is required")
	}
	ok, record, err := Verify(workbookRoot, evidenceID)
	if err != nil {
		return false, record, err
	}
	log, err := journal.Open(workbookRoot)
	if err != nil {
		return ok, record, err
	}
	event, err := journal.NewEvent("EVIDENCE_VERIFIED", "EVIDENCE", map[string]any{"workbook_id": workbookID, "evidence_id": record.ID, "verification": record.Verification, "size_bytes": record.Size, "sha256": record.SHA256}, now)
	if err != nil {
		return ok, record, err
	}
	if err := log.Append(event); err != nil {
		return ok, record, err
	}
	return ok, record, nil
}

func VerifyAllRecorded(workbookRoot, workbookID string, now func() time.Time) ([]Record, bool, error) {
	records, err := List(workbookRoot)
	if err != nil {
		return nil, false, err
	}
	verified := make([]Record, 0, len(records))
	allMatch := true
	for _, record := range records {
		ok, result, err := VerifyRecorded(workbookRoot, workbookID, record.ID, now())
		if err != nil {
			return nil, false, err
		}
		verified = append(verified, result)
		if !ok {
			allMatch = false
		}
	}
	return verified, allMatch, nil
}
