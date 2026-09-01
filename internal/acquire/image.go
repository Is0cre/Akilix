package acquire

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Is0cre/Akilix/internal/journal"
)

const ImageSchema = "akilix.acquisition-image.v1"

type ImageRecord struct {
	Schema      string    `json:"schema"`
	OperationID string    `json:"operation_id"`
	WorkbookID  string    `json:"workbook_id"`
	Phase       string    `json:"phase"`
	RecordedAt  time.Time `json:"recorded_at"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	SizeBytes   int64     `json:"size_bytes,omitempty"`
	SHA256      string    `json:"sha256,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// Image streams one already-inspected, unmounted, kernel-read-only whole disk
// into the workbook. It never mounts, unmounts, elevates, or changes the source.
func Image(ctx context.Context, workbookRoot, workbookID string, device Device, outputName string, now time.Time) (ImageRecord, string, error) {
	if workbookID == "" || device.Path == "" || device.SystemDisk || !device.AcquisitionCandidate || device.Mounted || !device.ReadOnly {
		return ImageRecord{}, "", fmt.Errorf("source must be an inspected, unmounted, kernel-read-only acquisition candidate")
	}
	if !validImageName(outputName) {
		return ImageRecord{}, "", fmt.Errorf("invalid image output name")
	}
	source, err := os.OpenFile(device.Path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return ImageRecord{}, "", fmt.Errorf("open source read-only: %w", err)
	}
	defer source.Close()
	info, err := source.Stat()
	if err != nil || info.Mode()&os.ModeDevice == 0 || info.Mode()&os.ModeCharDevice != 0 {
		return ImageRecord{}, "", fmt.Errorf("source is not a block device")
	}
	return imageFromReader(ctx, workbookRoot, workbookID, device.Path, outputName, source, int64(device.SizeBytes), now)
}

func imageFromReader(ctx context.Context, workbookRoot, workbookID, sourcePath, outputName string, source io.Reader, expectedSize int64, now time.Time) (ImageRecord, string, error) {
	if !validImageName(outputName) {
		return ImageRecord{}, "", fmt.Errorf("invalid image output name")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id, err := uuid7(now)
	if err != nil {
		return ImageRecord{}, "", err
	}
	dir := filepath.Join(workbookRoot, "evidence", "acquired")
	if err := ensureRealDirectory(filepath.Join(workbookRoot, "evidence")); err != nil {
		return ImageRecord{}, "", err
	}
	if err := ensureRealDirectory(dir); err != nil {
		return ImageRecord{}, "", err
	}
	destination := filepath.Join(dir, outputName)
	if _, err := os.Lstat(destination); err == nil {
		return ImageRecord{}, "", fmt.Errorf("acquisition destination already exists")
	} else if !os.IsNotExist(err) {
		return ImageRecord{}, "", err
	}
	requested := ImageRecord{Schema: ImageSchema, OperationID: id, WorkbookID: workbookID, Phase: "REQUESTED", RecordedAt: now.UTC(), Source: sourcePath, Destination: destination}
	if err := recordImage(workbookRoot, requested); err != nil {
		return ImageRecord{}, "", err
	}
	log, err := journal.Open(workbookRoot)
	if err != nil {
		return ImageRecord{}, "", err
	}
	if err := appendImageJournal(log, requested); err != nil {
		return ImageRecord{}, "", err
	}

	tmp, err := os.CreateTemp(dir, ".acquire-")
	if err != nil {
		return failImage(workbookRoot, log, requested, err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return failImage(workbookRoot, log, requested, err)
	}
	hash := sha256.New()
	buffer := make([]byte, 1024*1024)
	n, copyErr := copyContext(ctx, io.MultiWriter(tmp, hash), source, buffer)
	if copyErr == nil && expectedSize > 0 && n != expectedSize {
		copyErr = fmt.Errorf("source size changed or acquisition was incomplete: expected %d bytes, read %d", expectedSize, n)
	}
	if copyErr == nil {
		copyErr = tmp.Sync()
	}
	if closeErr := tmp.Close(); copyErr == nil {
		copyErr = closeErr
	}
	if copyErr == nil {
		copyErr = os.Link(tmpPath, destination)
	}
	if copyErr == nil {
		copyErr = os.Remove(tmpPath)
	}
	if copyErr != nil {
		return failImage(workbookRoot, log, requested, copyErr)
	}
	if err := syncDirectory(dir); err != nil {
		_ = os.Remove(destination)
		return failImage(workbookRoot, log, requested, err)
	}
	complete := requested
	complete.Phase, complete.RecordedAt, complete.SizeBytes, complete.SHA256 = "COMPLETED", time.Now().UTC(), n, hex.EncodeToString(hash.Sum(nil))
	if err := recordImage(workbookRoot, complete); err != nil {
		return complete, destination, err
	}
	if err := appendImageJournal(log, complete); err != nil {
		return complete, destination, err
	}
	return complete, destination, nil
}

func copyContext(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		n, readErr := src.Read(buf)
		if n > 0 {
			written, err := dst.Write(buf[:n])
			total += int64(written)
			if err != nil {
				return total, err
			}
			if written != n {
				return total, io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			return total, nil
		}
		if readErr != nil {
			return total, readErr
		}
	}
}

func failImage(root string, log *journal.Journal, requested ImageRecord, cause error) (ImageRecord, string, error) {
	failed := requested
	failed.Phase, failed.RecordedAt, failed.Error = "FAILED", time.Now().UTC(), cause.Error()
	if err := recordImage(root, failed); err != nil {
		return failed, "", fmt.Errorf("%v; record failure: %w", cause, err)
	}
	if err := appendImageJournal(log, failed); err != nil {
		return failed, "", fmt.Errorf("%v; journal failure: %w", cause, err)
	}
	return failed, "", cause
}

func recordImage(root string, record ImageRecord) error {
	dir := filepath.Join(root, "hardware", "acquisitions")
	if err := ensureRealDirectory(filepath.Join(root, "hardware")); err != nil {
		return err
	}
	if err := ensureRealDirectory(dir); err != nil {
		return err
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return writeNewAtomic(filepath.Join(dir, record.OperationID+"-"+strings.ToLower(record.Phase)+".json"), append(data, '\n'))
}

func appendImageJournal(log *journal.Journal, record ImageRecord) error {
	event, err := journal.NewEvent("ACQUISITION_IMAGE_"+record.Phase, "HARDWARE", map[string]any{"operation_id": record.OperationID, "source": record.Source, "destination": record.Destination, "size_bytes": record.SizeBytes, "sha256": record.SHA256, "error": record.Error}, record.RecordedAt)
	if err != nil {
		return err
	}
	return log.Append(event)
}

func validImageName(name string) bool {
	return name != "" && name == filepath.Base(name) && name != "." && name != ".." && !strings.ContainsAny(name, "/\\")
}
func syncDirectory(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
