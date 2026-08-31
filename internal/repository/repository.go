package repository

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

const Schema = "pensuse.repositories.v1"

type Set struct {
	Schema       string `json:"schema"`
	Repositories []Item `json:"repositories"`
}

type Item struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Purpose        string `json:"purpose"`
	Tier           string `json:"tier"`
	BaseURL        string `json:"base_url"`
	KeyURL         string `json:"key_url"`
	KeyFingerprint string `json:"key_fingerprint"`
	ImageEnabled   bool   `json:"image_enabled"`
	Status         string `json:"status"`
}

func Load(path string) (Set, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Set{}, err
	}
	var set Set
	if err := json.Unmarshal(b, &set); err != nil {
		return Set{}, err
	}
	return set, set.Validate()
}

func (s Set) Validate() error {
	if s.Schema != Schema || len(s.Repositories) == 0 {
		return fmt.Errorf("invalid repository set")
	}
	seen := map[string]bool{}
	for _, item := range s.Repositories {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("repository %q: %w", item.ID, err)
		}
		if seen[item.ID] {
			return fmt.Errorf("duplicate repository ID %q", item.ID)
		}
		seen[item.ID] = true
	}
	return nil
}

func (i Item) Validate() error {
	if i.ID == "" || strings.TrimSpace(i.Name) == "" || (i.Purpose != "base" && i.Purpose != "desktop" && i.Purpose != "boot") || (i.Tier != "release" && i.Tier != "obs") || (i.Status != "approved" && i.Status != "candidate") {
		return fmt.Errorf("invalid repository metadata")
	}
	if i.ImageEnabled && i.Status != "approved" {
		return fmt.Errorf("candidate repository cannot be image-enabled")
	}
	for _, raw := range []string{i.BaseURL, i.KeyURL} {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("repository URLs must use HTTPS")
		}
	}
	if len(i.KeyFingerprint) != 40 {
		return fmt.Errorf("invalid signing-key fingerprint")
	}
	for _, character := range i.KeyFingerprint {
		if !((character >= '0' && character <= '9') || (character >= 'A' && character <= 'F')) {
			return fmt.Errorf("invalid signing-key fingerprint")
		}
	}
	return nil
}

func (s Set) Find(id string) (Item, error) {
	for _, item := range s.Repositories {
		if item.ID == id {
			return item, nil
		}
	}
	return Item{}, fmt.Errorf("repository %q not found", id)
}
