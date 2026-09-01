package profile

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type hostCall struct {
	name string
	args []string
}

type fakeHost struct {
	missing    map[string]bool
	fs         string
	fsErr      error
	snapper    string
	snapperErr error
	euid       int
	calls      *[]hostCall
	lookups    *[]string
}

func (h fakeHost) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	if h.calls != nil {
		*h.calls = append(*h.calls, hostCall{name: name, args: append([]string(nil), args...)})
	}
	if name == "findmnt" {
		return []byte(h.fs), h.fsErr
	}
	if name == "snapper" {
		return []byte(h.snapper), h.snapperErr
	}
	return nil, errors.New("unexpected command")
}
func (h fakeHost) LookPath(name string) (string, error) {
	if h.lookups != nil {
		*h.lookups = append(*h.lookups, name)
	}
	if h.missing[name] {
		return "", errors.New("missing")
	}
	return "/usr/bin/" + name, nil
}
func (h fakeHost) EUID() int { return h.euid }

func TestPreflightHostReportsReadyFoundationWithoutApplying(t *testing.T) {
	m := Manifest{Schema: Schema, ID: "CORE", Name: "Core", Description: "Core", Status: "foundation"}
	p, err := PreflightHost(context.Background(), m, fakeHost{fs: "btrfs\n", snapper: "root /\n", euid: 1000})
	if err != nil || !p.Ready || p.ApplySupported || !p.RequiresPrivilege {
		t.Fatalf("preflight=%+v err=%v", p, err)
	}
}

func TestPreflightHostUsesOnlyReadOnlyInspectionCommands(t *testing.T) {
	var calls []hostCall
	m := Manifest{Schema: Schema, ID: "CORE", Name: "Core", Description: "Core", Status: "foundation"}
	_, err := PreflightHost(context.Background(), m, fakeHost{fs: "btrfs\n", snapper: "root /\n", calls: &calls})
	if err != nil {
		t.Fatal(err)
	}
	want := []hostCall{
		{name: "findmnt", args: []string{"-n", "-o", "FSTYPE", "/"}},
		{name: "snapper", args: []string{"--no-dbus", "list-configs", "--columns", "config,subvolume"}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("commands=%#v want=%#v", calls, want)
	}
}

func TestPreflightHostRejectsPlannedOrIncompleteHost(t *testing.T) {
	m := Manifest{Schema: Schema, ID: "NETWORK", Name: "Network", Description: "Network", Status: "planned"}
	p, err := PreflightHost(context.Background(), m, fakeHost{missing: map[string]bool{"snapper": true}, fs: "ext4\n"})
	if err != nil || p.Ready {
		t.Fatalf("preflight=%+v err=%v", p, err)
	}
}

func TestPreflightHostReportsLookupAndInspectionFailures(t *testing.T) {
	m := Manifest{Schema: Schema, ID: "CORE", Name: "Core", Description: "Core", Status: "foundation"}
	p, err := PreflightHost(context.Background(), m, fakeHost{
		missing:    map[string]bool{"findmnt": true},
		fs:         "findmnt diagnostic\n",
		fsErr:      errors.New("findmnt failed"),
		snapper:    "snapper diagnostic\n",
		snapperErr: errors.New("snapper failed"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Ready {
		t.Fatalf("preflight unexpectedly ready: %+v", p)
	}
	checks := make(map[string]PreflightCheck, len(p.Checks))
	for _, check := range p.Checks {
		checks[check.Name] = check
	}
	if got := checks["tool:findmnt"]; got.Ready || got.Detail != "required host tool unavailable" || got.Error != "missing" {
		t.Fatalf("findmnt lookup check=%+v", got)
	}
	if got := checks["root-filesystem"]; got.Ready || got.Detail != "unable to inspect root filesystem" || got.Error != "findmnt diagnostic" {
		t.Fatalf("filesystem check=%+v", got)
	}
	if got := checks["snapper-root-config"]; got.Ready || got.Detail != "unable to inspect Snapper configurations" || got.Error != "snapper diagnostic" {
		t.Fatalf("snapper check=%+v", got)
	}
}

func TestPreflightHostReportsCommandErrorsWithoutOutput(t *testing.T) {
	m := Manifest{Schema: Schema, ID: "CORE", Name: "Core", Description: "Core", Status: "foundation"}
	p, err := PreflightHost(context.Background(), m, fakeHost{fsErr: errors.New("findmnt unavailable"), snapperErr: errors.New("snapper unavailable")})
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Checks[5].Error; got != "findmnt unavailable" {
		t.Fatalf("filesystem error=%q", got)
	}
	if got := p.Checks[6].Error; got != "snapper unavailable" {
		t.Fatalf("snapper error=%q", got)
	}
}

func TestPreflightHostRejectsInvalidManifestBeforeHostInspection(t *testing.T) {
	var calls []hostCall
	var lookups []string
	p, err := PreflightHost(context.Background(), Manifest{}, fakeHost{calls: &calls, lookups: &lookups})
	if err == nil {
		t.Fatalf("preflight=%+v expected manifest error", p)
	}
	if len(calls) != 0 || len(lookups) != 0 {
		t.Fatalf("host inspected for invalid manifest: commands=%#v lookups=%#v", calls, lookups)
	}
}
