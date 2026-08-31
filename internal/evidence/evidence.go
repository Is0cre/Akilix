package evidence

import (
	"crypto/rand"
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
)

const Schema = "pensuse.evidence.v1"

type Record struct {
	Schema         string    `json:"schema"`
	ID             string    `json:"evidence_id"`
	Classification string    `json:"classification"`
	Filename       string    `json:"filename"`
	Source         string    `json:"source"`
	Size           int64     `json:"size"`
	SHA256         string    `json:"sha256"`
	Imported       time.Time `json:"imported"`
	Status         string    `json:"status"`
	Verification   string    `json:"verification,omitempty"`
}

func (r Record) Validate() error {
	if r.Schema != Schema || !validID(r.ID) || r.Filename == "" || filepath.Base(r.Filename) != r.Filename || r.Filename == "." || r.Filename == ".." || r.Size < 0 || !validHex(r.SHA256, 64) || r.Source == "" || r.Classification != "original" || r.Status != "complete" || (r.Verification != "" && r.Verification != "match" && r.Verification != "mismatch") {
		return fmt.Errorf("invalid evidence record")
	}
	return nil
}

func Import(workbookRoot, source string, now time.Time) (Record, error) {
	in, info, err := openRegularNoFollow(source)
	if err != nil {
		return Record{}, err
	}
	defer in.Close()
	name := filepath.Base(source)
	if name == "." || name == ".." || strings.ContainsAny(name, "/\\") {
		return Record{}, fmt.Errorf("invalid evidence filename")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id, err := uuid(now)
	if err != nil {
		return Record{}, err
	}
	dir := filepath.Join(workbookRoot, "evidence", "original")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return Record{}, err
	}
	if err := requireRealDirectory(filepath.Join(workbookRoot, "evidence")); err != nil {
		return Record{}, err
	}
	if err := requireRealDirectory(dir); err != nil {
		return Record{}, err
	}
	dest := filepath.Join(dir, name)
	if _, err := os.Lstat(dest); err == nil {
		return Record{}, fmt.Errorf("original evidence %q already exists", name)
	} else if !os.IsNotExist(err) {
		return Record{}, err
	}
	tmp, err := os.CreateTemp(dir, ".import-")
	if err != nil {
		return Record{}, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err = tmp.Chmod(0600); err != nil {
		tmp.Close()
		return Record{}, err
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), in)
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return Record{}, err
	}
	if err = os.Link(tmpName, dest); err != nil {
		return Record{}, err
	}
	if err = os.Remove(tmpName); err != nil {
		_ = os.Remove(dest)
		return Record{}, err
	}
	if latest, statErr := in.Stat(); statErr != nil || latest.Size() != info.Size() || !latest.ModTime().Equal(info.ModTime()) {
		_ = os.Remove(dest)
		if statErr != nil {
			return Record{}, fmt.Errorf("evidence source changed or became unavailable: %w", statErr)
		}
		return Record{}, fmt.Errorf("evidence source changed during import")
	}
	if err := syncDir(dir); err != nil {
		_ = os.Remove(dest)
		return Record{}, err
	}
	record := Record{Schema: Schema, ID: id, Classification: "original", Filename: name, Source: filepath.Clean(source), Size: n, SHA256: hex.EncodeToString(h.Sum(nil)), Imported: now.UTC(), Status: "complete"}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return Record{}, err
	}
	data = append(data, '\n')
	manifest := filepath.Join(workbookRoot, "evidence", "manifests", id+".json")
	if err := os.MkdirAll(filepath.Dir(manifest), 0700); err != nil {
		_ = os.Remove(dest)
		return Record{}, err
	}
	if err := requireRealDirectory(filepath.Dir(manifest)); err != nil {
		_ = os.Remove(dest)
		return Record{}, err
	}
	if err := atomicWrite(manifest, data); err != nil {
		_ = os.Remove(dest)
		return Record{}, err
	}
	return record, nil
}

func Verify(workbookRoot, id string) (bool, Record, error) {
	if !validID(id) {
		return false, Record{}, fmt.Errorf("invalid evidence ID")
	}
	if err := requireEvidenceDirectories(workbookRoot); err != nil {
		return false, Record{}, err
	}
	b, err := readRegularNoFollow(filepath.Join(workbookRoot, "evidence", "manifests", id+".json"))
	if err != nil {
		return false, Record{}, err
	}
	var r Record
	if err := json.Unmarshal(b, &r); err != nil {
		return false, Record{}, err
	}
	if err := r.Validate(); err != nil {
		return false, Record{}, err
	}
	if r.ID != id {
		return false, r, fmt.Errorf("evidence manifest ID does not match requested ID")
	}
	f, _, err := openRegularNoFollow(filepath.Join(workbookRoot, "evidence", "original", r.Filename))
	if err != nil {
		return false, r, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return false, r, err
	}
	ok := n == r.Size && hex.EncodeToString(h.Sum(nil)) == r.SHA256
	if ok {
		r.Verification = "match"
	} else {
		r.Verification = "mismatch"
	}
	data, marshalErr := json.MarshalIndent(r, "", "  ")
	if marshalErr != nil {
		return ok, r, marshalErr
	}
	data = append(data, '\n')
	if writeErr := atomicWrite(filepath.Join(workbookRoot, "evidence", "manifests", id+".json"), data); writeErr != nil {
		return ok, r, writeErr
	}
	return ok, r, nil
}

// VerifyAll verifies every canonical evidence manifest in deterministic order.
func VerifyAll(workbookRoot string) ([]Record, bool, error) {
	records, err := List(workbookRoot)
	if err != nil {
		return nil, false, err
	}
	allMatch := true
	verified := make([]Record, 0, len(records))
	for _, record := range records {
		ok, updated, err := Verify(workbookRoot, record.ID)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			allMatch = false
		}
		verified = append(verified, updated)
	}
	return verified, allMatch, nil
}

// CheckAll validates every evidence manifest and hashes each original without
// modifying manifests. It is suitable for read-only workbook audits.
func CheckAll(workbookRoot string) ([]Record, bool, error) {
	records, err := List(workbookRoot)
	if err != nil {
		return nil, false, err
	}
	allMatch := true
	for i := range records {
		ok, err := checkRecord(workbookRoot, records[i])
		if err != nil {
			return nil, false, err
		}
		if !ok {
			allMatch = false
		}
		if ok {
			records[i].Verification = "match"
		} else {
			records[i].Verification = "mismatch"
		}
	}
	return records, allMatch, nil
}

func checkRecord(workbookRoot string, r Record) (bool, error) {
	f, _, err := openRegularNoFollow(filepath.Join(workbookRoot, "evidence", "original", r.Filename))
	if err != nil {
		return false, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return false, err
	}
	return n == r.Size && hex.EncodeToString(h.Sum(nil)) == r.SHA256, nil
}

func List(workbookRoot string) ([]Record, error) {
	manifestDir := filepath.Join(workbookRoot, "evidence", "manifests")
	if err := requireRealDirectory(manifestDir); err != nil {
		if os.IsNotExist(err) {
			return []Record{}, nil
		}
		return nil, err
	}
	for _, path := range []string{filepath.Join(workbookRoot, "evidence"), filepath.Join(workbookRoot, "evidence", "original")} {
		if err := requireRealDirectory(path); err != nil {
			return nil, err
		}
	}
	entries, err := os.ReadDir(manifestDir)
	if err != nil {
		return nil, err
	}
	out := make([]Record, 0, len(entries))
	for _, e := range entries {
		if e.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("evidence manifest must not be a symlink: %q", e.Name())
		}
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := readRegularNoFollow(filepath.Join(workbookRoot, "evidence", "manifests", e.Name()))
		if err != nil {
			return nil, err
		}
		var r Record
		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
		if err := r.Validate(); err != nil {
			return nil, err
		}
		if strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())) != r.ID {
			return nil, fmt.Errorf("evidence manifest filename does not match record ID")
		}
		out = append(out, r)
	}
	return out, nil
}

func readRegularNoFollow(path string) ([]byte, error) {
	f, _, err := openRegularNoFollow(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func openRegularNoFollow(path string) (*os.File, os.FileInfo, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		f.Close()
		return nil, nil, fmt.Errorf("evidence source must be a regular file")
	}
	return f, info, nil
}

func requireRealDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("evidence path must be a real directory: %q", path)
	}
	return nil
}

func requireEvidenceDirectories(workbookRoot string) error {
	for _, path := range []string{
		filepath.Join(workbookRoot, "evidence"),
		filepath.Join(workbookRoot, "evidence", "original"),
		filepath.Join(workbookRoot, "evidence", "manifests"),
	} {
		if err := requireRealDirectory(path); err != nil {
			return err
		}
	}
	return nil
}

func uuid(t time.Time) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[6:]); err != nil {
		return "", err
	}
	ms := uint64(t.UnixMilli())
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	b[6] = (b[6] & 15) | 112
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:])), nil
}
func atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".manifest-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err == nil {
		err = syncDir(filepath.Dir(path))
	}
	return err
}

func syncDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

func validHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, c := range value {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func validID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || value[14] != '7' {
		return false
	}
	variant := value[19]
	validVariant := (variant >= '8' && variant <= '9') || (variant >= 'a' && variant <= 'b') || (variant >= 'A' && variant <= 'B')
	if !validVariant {
		return false
	}
	for i, c := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
