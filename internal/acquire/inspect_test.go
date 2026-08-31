package acquire

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeRunner struct {
	name string
	args []string
	out  []byte
	err  error
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.name = name
	f.args = append([]string(nil), args...)
	return f.out, f.err
}

func TestInspectClassifiesSystemAndCandidateDisks(t *testing.T) {
	runner := &fakeRunner{out: []byte(`{"blockdevices":[{"name":"sdb","kname":"sdb","path":"/dev/sdb","type":"disk","size":2000,"ro":true,"rm":true,"tran":"usb","vendor":"ACME ","model":"Evidence","serial":"SER2","wwn":"","mountpoints":[null],"children":[{"name":"sdb1","path":"/dev/sdb1","type":"part","size":1900,"ro":true,"fstype":"exfat","uuid":"U-2","mountpoints":[null]}]},{"name":"nvme0n1","kname":"nvme0n1","path":"/dev/nvme0n1","type":"disk","size":1000,"ro":false,"rm":false,"tran":"nvme","mountpoints":[null],"children":[{"name":"nvme0n1p2","path":"/dev/nvme0n1p2","type":"part","size":900,"ro":false,"fstype":"btrfs","uuid":"U-1","mountpoints":["/"]}]}]}`)}
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	report, err := Inspect(context.Background(), runner, now)
	if err != nil {
		t.Fatal(err)
	}
	if runner.name != "lsblk" || strings.Join(runner.args, " ") != "--json --bytes --tree --output "+lsblkColumns {
		t.Fatalf("unexpected command: %s %v", runner.name, runner.args)
	}
	if report.Schema != InspectionSchema || !report.Passive || report.GeneratedAt != now || len(report.Devices) != 2 {
		t.Fatalf("unexpected report: %+v", report)
	}
	system, candidate := report.Devices[0], report.Devices[1]
	if !system.SystemDisk || system.AcquisitionCandidate || !system.Mounted {
		t.Fatalf("system classification: %+v", system)
	}
	if candidate.SystemDisk || !candidate.AcquisitionCandidate || !candidate.ReadOnly || candidate.Transport != "usb" || candidate.Vendor != "ACME" || len(candidate.Partitions) != 1 {
		t.Fatalf("candidate classification: %+v", candidate)
	}
}

func TestInspectRejectsInvalidJSONAndPropagatesCommandFailure(t *testing.T) {
	if _, err := Inspect(context.Background(), &fakeRunner{out: []byte("no")}, time.Now()); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	if _, err := Inspect(context.Background(), &fakeRunner{err: errors.New("missing")}, time.Now()); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("command failure: %v", err)
	}
}
