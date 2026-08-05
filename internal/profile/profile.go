package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const Schema = "pensuse.profile.v1"

type Manifest struct {
	Schema      string   `json:"schema"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	RPM         []string `json:"rpm"`
	Containers  []string `json:"containers"`
}

func (m Manifest) Validate() error {
	if m.Schema != Schema || !validID(m.ID) || strings.TrimSpace(m.Name) == "" || strings.TrimSpace(m.Description) == "" {
		return fmt.Errorf("invalid profile manifest")
	}
	if m.Status != "foundation" && m.Status != "planned" {
		return fmt.Errorf("invalid profile status %q", m.Status)
	}
	seen := map[string]bool{}
	for _, item := range append(append([]string{}, m.RPM...), m.Containers...) {
		if strings.TrimSpace(item) == "" || seen[item] {
			return fmt.Errorf("invalid or duplicate profile component %q", item)
		}
		seen[item] = true
	}
	return nil
}

func LoadDir(dir string) ([]Manifest, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []Manifest{}, nil
	}
	if err != nil {
		return nil, err
	}
	manifests := make([]Manifest, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		m, err := load(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("profile %q: %w", entry.Name(), err)
		}
		manifests = append(manifests, m)
	}
	sort.Slice(manifests, func(i, j int) bool { return manifests[i].ID < manifests[j].ID })
	return manifests, nil
}

func Find(dir, id string) (Manifest, error) {
	if !validID(id) {
		return Manifest{}, fmt.Errorf("invalid profile ID")
	}
	items, err := LoadDir(dir)
	if err != nil {
		return Manifest{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return Manifest{}, fmt.Errorf("profile %q not found", id)
}

func load(path string) (Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	section := ""
	for _, line := range strings.Split(string(b), "\n") {
		s := strings.TrimSpace(line)
		if s == "rpm:" || s == "containers:" {
			section = strings.TrimSuffix(s, ":")
			continue
		}
		if strings.HasPrefix(s, "- ") {
			item := strings.TrimSpace(strings.TrimPrefix(s, "- "))
			if section == "rpm" {
				m.RPM = append(m.RPM, item)
			} else if section == "containers" {
				m.Containers = append(m.Containers, item)
			}
			continue
		}
		key, value, ok := strings.Cut(s, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch key {
		case "schema":
			m.Schema = value
		case "id":
			m.ID = value
		case "name":
			m.Name = value
		case "description":
			m.Description = value
		case "status":
			m.Status = value
		}
	}
	if err := m.Validate(); err != nil {
		return Manifest{}, err
	}
	if m.RPM == nil { m.RPM = []string{} }
	if m.Containers == nil { m.Containers = []string{} }
	return m, nil
}

func validID(id string) bool {
	if id == "" || id != strings.ToUpper(id) || strings.ContainsAny(id, "/\\ ") {
		return false
	}
	for _, c := range id {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}
