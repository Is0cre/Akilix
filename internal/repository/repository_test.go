package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndCandidateCannotBeEnabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "repositories.json")
	data := `{"schema":"akilix.repositories.v1","repositories":[{"id":"leap-oss","name":"Leap OSS","purpose":"base","tier":"release","base_url":"https://example.test/repo/","key_url":"https://example.test/key","key_fingerprint":"AD485664E901B867051AB15F35A2F86E29B700A4","image_enabled":true,"status":"approved"}]}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	set, err := Load(path)
	if err != nil || len(set.Repositories) != 1 {
		t.Fatalf("load: %+v %v", set, err)
	}
	item := set.Repositories[0]
	item.Status = "candidate"
	if err := item.Validate(); err == nil {
		t.Fatal("image-enabled candidate repository accepted")
	}
}

func TestRejectsInsecureRepositoryURL(t *testing.T) {
	item := Item{ID: "test", Name: "Test", Purpose: "base", Tier: "release", BaseURL: "http://example.test", KeyURL: "https://example.test/key", KeyFingerprint: "AD485664E901B867051AB15F35A2F86E29B700A4", Status: "approved"}
	if err := item.Validate(); err == nil {
		t.Fatal("HTTP repository accepted")
	}
}

func TestAcceptsAuditedSecurityRepository(t *testing.T) {
	item := Item{ID: "security-leap-16", Name: "Security", Purpose: "security", Tier: "obs", BaseURL: "https://download.opensuse.org/repositories/security/16.0/", KeyURL: "https://download.opensuse.org/repositories/security/16.0/repodata/repomd.xml.key", KeyFingerprint: "F9FA0223B56B116C363737EF5DA57BDD6DD785CA", ImageEnabled: true, Status: "approved"}
	if err := item.Validate(); err != nil {
		t.Fatal(err)
	}
}
