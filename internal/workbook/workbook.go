package workbook

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const Schema = "pensuse.workbook.v1"

type Metadata struct {
	Schema  string
	ID      string
	Name    string
	Created time.Time
	Status  string
}

func (m Metadata) Validate() error {
	if m.Schema != Schema || m.ID == "" || m.Name == "" || m.Created.IsZero() || (m.Status != "open" && m.Status != "closed") {
		return fmt.Errorf("invalid workbook metadata")
	}
	return nil
}

func Create(root, name string, now time.Time) (Metadata, error) {
	if err := validName(name); err != nil {
		return Metadata{}, err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id, err := uuid7(now)
	if err != nil {
		return Metadata{}, err
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return Metadata{}, err
	}
	if _, err := os.Lstat(filepath.Join(root, name)); err == nil {
		return Metadata{}, fmt.Errorf("workbook %q already exists", name)
	} else if !os.IsNotExist(err) {
		return Metadata{}, err
	}
	stage, err := os.MkdirTemp(root, ".creating-")
	if err != nil {
		return Metadata{}, err
	}
	defer os.RemoveAll(stage)
	m := Metadata{Schema: Schema, ID: id, Name: name, Created: now.UTC(), Status: "open"}
	for _, sub := range []string{"evidence/original", "evidence/acquired", "evidence/manifests", "artifacts/imported", "artifacts/derived", "artifacts/extracted", "captures", "tool-output", "notes", "findings", "timeline", "reports/drafts", "reports/exports", "logs/command", "logs/containers", "logs/audit", ".pensuse/locks"} {
		if err := os.MkdirAll(filepath.Join(stage, sub), 0700); err != nil {
			return Metadata{}, err
		}
	}
	if err := atomicWrite(filepath.Join(stage, "workbook.yaml"), render(m), 0600); err != nil {
		return Metadata{}, err
	}
	if err := atomicWrite(filepath.Join(stage, "scope.yaml"), "version: 1\ninclude:\n  []\nexclude:\n  []\n", 0600); err != nil {
		return Metadata{}, err
	}
	if err := atomicWrite(filepath.Join(stage, "README.md"), "# "+name+"\n\nPenSUSE workbook.\n", 0600); err != nil {
		return Metadata{}, err
	}
	if err := syncDir(stage); err != nil {
		return Metadata{}, err
	}
	if err := os.Rename(stage, filepath.Join(root, name)); err != nil {
		if errors.Is(err, os.ErrExist) {
			return Metadata{}, fmt.Errorf("workbook %q already exists", name)
		}
		return Metadata{}, err
	}
	if err := syncDir(root); err != nil {
		return Metadata{}, err
	}
	return m, nil
}

func Open(root, name string) (Metadata, error) {
	if err := validName(name); err != nil {
		return Metadata{}, err
	}
	workbookDir := filepath.Join(root, name)
	if err := requirePath(workbookDir, true); err != nil {
		return Metadata{}, err
	}
	metadataPath := filepath.Join(workbookDir, "workbook.yaml")
	if err := requirePath(metadataPath, false); err != nil {
		return Metadata{}, err
	}
	b, err := os.ReadFile(metadataPath)
	if err != nil {
		return Metadata{}, err
	}
	m, err := parse(string(b))
	if err != nil {
		return Metadata{}, err
	}
	if m.Name != name {
		return Metadata{}, fmt.Errorf("workbook metadata name does not match directory")
	}
	return m, nil
}

func SetStatus(root, name, status string) (Metadata, error) {
	if status != "open" && status != "closed" {
		return Metadata{}, fmt.Errorf("invalid workbook status %q", status)
	}
	m, err := Open(root, name)
	if err != nil {
		return Metadata{}, err
	}
	m.Status = status
	if err := atomicWrite(filepath.Join(root, name, "workbook.yaml"), render(m), 0600); err != nil {
		return Metadata{}, err
	}
	return m, nil
}

func Rename(root, oldName, newName string) (Metadata, error) {
	if err := validName(oldName); err != nil {
		return Metadata{}, err
	}
	if err := validName(newName); err != nil {
		return Metadata{}, err
	}
	m, err := Open(root, oldName)
	if err != nil {
		return Metadata{}, err
	}
	if _, err := os.Lstat(filepath.Join(root, newName)); err == nil {
		return Metadata{}, fmt.Errorf("workbook %q already exists", newName)
	} else if !os.IsNotExist(err) {
		return Metadata{}, err
	}
	oldDir, newDir := filepath.Join(root, oldName), filepath.Join(root, newName)
	if err := os.Rename(oldDir, newDir); err != nil {
		return Metadata{}, err
	}
	m.Name = newName
	if err := atomicWrite(filepath.Join(newDir, "workbook.yaml"), render(m), 0600); err != nil {
		_ = os.Rename(newDir, oldDir)
		return Metadata{}, err
	}
	return m, nil
}

func List(root string) ([]Metadata, error) {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return []Metadata{}, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Metadata
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".creating-") {
			m, err := Open(root, e.Name())
			if err != nil {
				return nil, fmt.Errorf("invalid workbook %q: %w", e.Name(), err)
			}
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// ValidateLayout checks the canonical workbook files and directories without
// consulting any optional index database.
func ValidateLayout(root, name string) error {
	if err := validName(name); err != nil {
		return err
	}
	if _, err := Open(root, name); err != nil {
		return err
	}
	dir := filepath.Join(root, name)
	for _, file := range []string{"workbook.yaml", "scope.yaml", "README.md"} {
		if err := requirePath(filepath.Join(dir, file), false); err != nil {
			return err
		}
	}
	for _, sub := range []string{"evidence/original", "evidence/acquired", "evidence/manifests", "artifacts/imported", "artifacts/derived", "artifacts/extracted", "captures", "tool-output", "notes", "findings", "timeline", "reports/drafts", "reports/exports", "logs/command", "logs/containers", "logs/audit", ".pensuse/locks"} {
		if err := requirePath(filepath.Join(dir, sub), true); err != nil {
			return err
		}
	}
	return nil
}

func requirePath(path string, directory bool) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("missing workbook path %q: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || info.IsDir() != directory {
		return fmt.Errorf("invalid workbook path %q", path)
	}
	return nil
}

func validName(name string) error {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name || strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("invalid workbook name")
	}
	return nil
}

func uuid7(t time.Time) (string, error) {
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
	b[6] = (b[6] & 0x0f) | 0x70
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16])), nil
}

func render(m Metadata) string {
	return fmt.Sprintf("schema: %s\nid: %s\nname: %s\ncreated: %s\nstatus: %s\n", m.Schema, m.ID, m.Name, m.Created.Format(time.RFC3339Nano), m.Status)
}
func parse(s string) (Metadata, error) {
	var m Metadata
	for _, line := range strings.Split(s, "\n") {
		p := strings.SplitN(line, ": ", 2)
		if len(p) != 2 {
			continue
		}
		switch p[0] {
		case "schema":
			m.Schema = p[1]
		case "id":
			m.ID = p[1]
		case "name":
			m.Name = p[1]
		case "created":
			t, err := time.Parse(time.RFC3339Nano, p[1])
			if err != nil {
				return Metadata{}, err
			}
			m.Created = t
		case "status":
			m.Status = p[1]
		}
	}
	if err := m.Validate(); err != nil {
		return Metadata{}, err
	}
	return m, nil
}
func atomicWrite(path, data string, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(mode); err == nil {
		_, err = tmp.WriteString(data)
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
