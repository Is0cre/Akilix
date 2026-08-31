package acquire

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTrustUsesStableIdentityAndNeverDevicePath(t *testing.T) {
	a := Device{Path: "/dev/sdb", Vendor: "Acme", Model: "Vault", Serial: "ABC"}
	b := Device{Path: "/dev/sdz", Vendor: "Acme", Model: "Vault", Serial: "ABC"}
	idA, identityA, err := TrustIdentity(a)
	if err != nil {
		t.Fatal(err)
	}
	idB, identityB, err := TrustIdentity(b)
	if err != nil || idA != idB || identityA != identityB {
		t.Fatalf("identity changed with node: %q/%q %q/%q err=%v", idA, idB, identityA, identityB, err)
	}
	if _, _, err := TrustIdentity(Device{Path: "/dev/sdb", Model: "Generic"}); err == nil {
		t.Fatal("weak identity accepted")
	}
}

func TestTrustRegistryRoundTripAddMatchRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devices", "trusted.json")
	device := Device{Path: "/dev/sdb", WWN: "0xABC", Serial: "S1"}
	registry, err := LoadTrust(path)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := registry.Add(device, "lab disk", time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Add(device, "duplicate", time.Now()); err == nil {
		t.Fatal("duplicate trust accepted")
	}
	if err := SaveTrust(path, registry); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("mode=%v err=%v", info.Mode().Perm(), err)
	}
	loaded, err := LoadTrust(path)
	if err != nil {
		t.Fatal(err)
	}
	if match, ok := loaded.Match(device); !ok || match.ID != entry.ID {
		t.Fatalf("match=%+v ok=%t", match, ok)
	}
	if _, err := loaded.Remove(entry.ID, time.Now()); err != nil || len(loaded.Entries) != 0 || len(loaded.Revocations) != 1 {
		t.Fatalf("remove err=%v entries=%+v revocations=%+v", err, loaded.Entries, loaded.Revocations)
	}
}

func TestApplyTrustOnlyLabelsMatchingIdentity(t *testing.T) {
	trusted := Device{Path: "/dev/sdb", WWN: "0xABC"}
	unknown := Device{Path: "/dev/sdc", WWN: "0xDEF"}
	registry := TrustRegistry{Schema: TrustSchema, Entries: []TrustEntry{}, Revocations: []TrustRevocation{}}
	entry, err := registry.Add(trusted, "known", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	report := Inspection{Devices: []Device{trusted, unknown}}
	ApplyTrust(&report, registry)
	if !report.Devices[0].Trusted || report.Devices[0].TrustID != entry.ID || report.Devices[1].Trusted || report.Devices[1].TrustID != "" {
		t.Fatalf("unexpected trust labels: %+v", report.Devices)
	}
	if report.Devices[0].ReadOnly != trusted.ReadOnly {
		t.Fatal("trust annotation changed device state")
	}
}

func TestSaveTrustRejectsSymlinkDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "devices")); err != nil {
		t.Fatal(err)
	}
	registry := TrustRegistry{Schema: TrustSchema, Entries: []TrustEntry{}, Revocations: []TrustRevocation{}}
	if err := SaveTrust(filepath.Join(root, "devices", "trusted.json"), registry); err == nil {
		t.Fatal("symlink directory accepted")
	}
}
