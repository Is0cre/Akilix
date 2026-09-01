package profile

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type verifyRunner struct{}

func (verifyRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	if name == "rpm" && args[len(args)-1] == "nmap" {
		return nil, nil
	}
	if name == "podman" && args[len(args)-1] == "localhost/akilix-network" {
		return []byte("sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n"), nil
	}
	return []byte("not installed"), errors.New("exit 1")
}

func TestVerifyIsLocalAndReportsMissingComponents(t *testing.T) {
	m := Manifest{Schema: Schema, ID: "NETWORK", Name: "Network", Description: "test", Status: "planned", RPM: []string{"nmap", "tcpdump"}, Containers: []string{"akilix-network"}}
	v, err := Verify(context.Background(), m, verifyRunner{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if v.Ready || len(v.Components) != 3 || !v.Components[0].Present || v.Components[1].Present || !v.Components[2].Present {
		t.Fatalf("verification=%+v", v)
	}
	path, err := RecordVerification(t.TempDir(), v)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("path=%s info=%v err=%v", path, info, err)
	}
}

func TestRecordVerificationRejectsSymlinkDirectory(t *testing.T) {
	root := t.TempDir()
	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "platform"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "platform", "profile-verifications")); err != nil {
		t.Fatal(err)
	}
	_, err := RecordVerification(root, Verification{Schema: VerificationSchema, ID: "x", ProfileID: "CORE", VerifiedAt: time.Now()})
	if err == nil {
		t.Fatal("symlink directory accepted")
	}
}

func TestVerificationHistoryIsSortedAndFilterable(t *testing.T) {
	root := t.TempDir()
	later := Verification{Schema: VerificationSchema, ID: "later", ProfileID: "NETWORK", VerifiedAt: time.Now().Add(time.Second), Ready: false, Components: []ComponentState{}}
	earlier := Verification{Schema: VerificationSchema, ID: "earlier", ProfileID: "CORE", VerifiedAt: time.Now(), Ready: true, Components: []ComponentState{}}
	if _, err := RecordVerification(root, later); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordVerification(root, earlier); err != nil {
		t.Fatal(err)
	}
	history, err := VerificationHistory(root, "")
	if err != nil || len(history) != 2 || history[0].ID != "earlier" {
		t.Fatalf("history=%+v err=%v", history, err)
	}
	history, err = VerificationHistory(root, "NETWORK")
	if err != nil || len(history) != 1 || history[0].ID != "later" {
		t.Fatalf("filtered=%+v err=%v", history, err)
	}
}
