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
	if r.Schema != Schema || r.ID == "" || r.Filename == "" || r.Size < 0 || len(r.SHA256) != 64 || r.Status == "" {
		return fmt.Errorf("invalid evidence record")
	}
	return nil
}

func Import(workbookRoot, source string, now time.Time) (Record, error) {
	info, err := os.Lstat(source)
	if err != nil {
		return Record{}, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return Record{}, fmt.Errorf("evidence source must be a regular file")
	}
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
	in, err := os.Open(source)
	if err != nil {
		tmp.Close()
		return Record{}, err
	}
	defer in.Close()
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
	if err = os.Rename(tmpName, dest); err != nil {
		return Record{}, err
	}
	if latest, statErr := os.Stat(source); statErr != nil || latest.Size() != info.Size() {
		_ = os.Remove(dest)
		if statErr != nil {
			return Record{}, fmt.Errorf("evidence source changed or became unavailable: %w", statErr)
		}
		return Record{}, fmt.Errorf("evidence source changed during import")
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
	if err := atomicWrite(manifest, data); err != nil {
		_ = os.Remove(dest)
		return Record{}, err
	}
	return record, nil
}

func Verify(workbookRoot, id string) (bool, Record, error) {
	b, err := os.ReadFile(filepath.Join(workbookRoot, "evidence", "manifests", id+".json"))
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
	f, err := os.Open(filepath.Join(workbookRoot, "evidence", "original", r.Filename))
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

func List(workbookRoot string) ([]Record, error) {
	entries, err := os.ReadDir(filepath.Join(workbookRoot, "evidence", "manifests"))
	if os.IsNotExist(err) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]Record, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(workbookRoot, "evidence", "manifests", e.Name()))
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
		out = append(out, r)
	}
	return out, nil
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
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	return err
}
