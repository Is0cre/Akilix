package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDirAndFind(t *testing.T) {
	dir := t.TempDir()
	data := "schema: akilix.profile.v1\nid: CORE\nname: Core Platform\ndescription: Minimal platform\nstatus: foundation\nrpm:\n  - zsh\ncontainers:\n  []\n"
	if err := os.WriteFile(filepath.Join(dir, "CORE.yaml"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	items, err := LoadDir(dir)
	if err != nil || len(items) != 1 {
		t.Fatalf("load: %+v %v", items, err)
	}
	got, err := Find(dir, "CORE")
	if err != nil || got.Name != "Core Platform" || len(got.RPM) != 1 {
		t.Fatalf("find: %+v %v", got, err)
	}
}

func TestManifestRejectsDuplicateComponents(t *testing.T) {
	m := Manifest{Schema: Schema, ID: "CORE", Name: "Core", Description: "Core", Status: "foundation", RPM: []string{"zsh", "zsh"}}
	if err := m.Validate(); err == nil {
		t.Fatal("duplicate components accepted")
	}
}

func TestBuildPlanHasTransactionalPhases(t *testing.T) {
	plan := BuildPlan(Manifest{ID: "CORE", RPM: []string{"zypper"}, Containers: []string{"recon"}})
	if len(plan.Steps) != 5 || plan.Steps[0].Phase != "SNAPSHOT" || plan.Steps[len(plan.Steps)-1].Phase != "ROLLBACK" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestShippedProfileCatalog(t *testing.T) {
	dir := filepath.Join("..", "..", "profiles")
	items, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"CORE": true, "FORENSICS": true, "INCIDENT-RESPONSE": true,
		"NETWORK": true, "OPERATOR": true, "RECON": true,
		"REVERSE": true, "WEB": true, "WIRELESS": true,
	}
	for _, item := range items {
		delete(want, item.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing shipped profiles: %v", want)
	}
}
