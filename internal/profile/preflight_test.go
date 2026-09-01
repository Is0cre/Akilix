package profile

import (
	"context"
	"errors"
	"testing"
)

type fakeHost struct {
	missing map[string]bool
	fs      string
	snapper string
	euid    int
}

func (h fakeHost) Run(_ context.Context, name string, _ ...string) ([]byte, error) {
	if name == "findmnt" {
		return []byte(h.fs), nil
	}
	if name == "snapper" {
		return []byte(h.snapper), nil
	}
	return nil, errors.New("unexpected command")
}
func (h fakeHost) LookPath(name string) (string, error) {
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

func TestPreflightHostRejectsPlannedOrIncompleteHost(t *testing.T) {
	m := Manifest{Schema: Schema, ID: "NETWORK", Name: "Network", Description: "Network", Status: "planned"}
	p, err := PreflightHost(context.Background(), m, fakeHost{missing: map[string]bool{"snapper": true}, fs: "ext4\n"})
	if err != nil || p.Ready {
		t.Fatalf("preflight=%+v err=%v", p, err)
	}
}
