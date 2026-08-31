package acquire

import (
	"context"
	"errors"
	"testing"
)

type call struct {
	name string
	args []string
}
type sequenceRunner struct {
	calls   []call
	outputs [][]byte
	errors  []error
}

func (r *sequenceRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, call{name, append([]string(nil), args...)})
	i := len(r.calls) - 1
	var out []byte
	var err error
	if i < len(r.outputs) {
		out = r.outputs[i]
	}
	if i < len(r.errors) {
		err = r.errors[i]
	}
	return out, err
}

func TestCandidateRejectsSystemMountedAndPartitionPaths(t *testing.T) {
	report := Inspection{Devices: []Device{{Path: "/dev/sda", SystemDisk: true}, {Path: "/dev/sdb", AcquisitionCandidate: true, Mounted: true}, {Path: "/dev/sdc", AcquisitionCandidate: true}}}
	for _, path := range []string{"/dev/sda", "/dev/sdb", "/dev/sdc1", "/dev/missing"} {
		if _, err := Candidate(report, path); err == nil {
			t.Fatalf("accepted %s", path)
		}
	}
	if got, err := Candidate(report, "/dev/sdc"); err != nil || got.Path != "/dev/sdc" {
		t.Fatalf("candidate=%+v err=%v", got, err)
	}
}

func TestSetReadOnlyUsesExactArgumentVectorsAndVerifies(t *testing.T) {
	runner := &sequenceRunner{outputs: [][]byte{nil, []byte("1\n")}}
	result, err := SetReadOnly(context.Background(), runner, Device{Path: "/dev/sdb", AcquisitionCandidate: true})
	if err != nil || !result.KernelReadOnly || len(runner.calls) != 2 {
		t.Fatalf("result=%+v calls=%+v err=%v", result, runner.calls, err)
	}
	if runner.calls[0].name != "blockdev" || runner.calls[0].args[0] != "--setro" || runner.calls[0].args[1] != "/dev/sdb" || runner.calls[1].args[0] != "--getro" {
		t.Fatalf("unexpected calls: %+v", runner.calls)
	}
}

func TestSetReadOnlyDoesNotSetAgainAndReportsFailures(t *testing.T) {
	runner := &sequenceRunner{outputs: [][]byte{[]byte("1\n")}}
	result, err := SetReadOnly(context.Background(), runner, Device{Path: "/dev/sdb", ReadOnly: true, AcquisitionCandidate: true})
	if err != nil || !result.AlreadyReadOnly || len(runner.calls) != 1 || runner.calls[0].args[0] != "--getro" {
		t.Fatalf("result=%+v calls=%+v err=%v", result, runner.calls, err)
	}
	failing := &sequenceRunner{errors: []error{errors.New("permission denied")}}
	if _, err := SetReadOnly(context.Background(), failing, Device{Path: "/dev/sdb", AcquisitionCandidate: true}); err == nil {
		t.Fatal("setro failure accepted")
	}
}
