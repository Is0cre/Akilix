package acquire

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type resultCall struct {
	name string
	args []string
}
type fakeResultRunner struct {
	calls []resultCall
	outs  [][]byte
	exits []int
	errs  []error
}

func TestRecordIdentificationCreatesImmutableWorkbookRecord(t *testing.T) {
	root := t.TempDir()
	identification := Identification{Schema: IdentificationSchema, InventoryRecordID: "inventory-id", Passive: true, Device: Device{Path: "/dev/sdb"}, Commands: []CommandResult{{Tool: "smartctl", Args: []string{"--json=c"}, ExitCode: 0, Status: "OK", Output: []byte(`{}`)}}}
	record, path, err := RecordIdentification(root, "workbook-id", identification, time.Now())
	if err != nil || record.ID == "" || filepath.Dir(path) != filepath.Join(root, "hardware", "identifications") {
		t.Fatalf("record=%+v path=%q err=%v", record, path, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("mode=%v err=%v", info.Mode().Perm(), err)
	}
}

func (r *fakeResultRunner) RunResult(_ context.Context, name string, args ...string) ([]byte, int, error) {
	i := len(r.calls)
	r.calls = append(r.calls, resultCall{name, append([]string(nil), args...)})
	var out []byte
	var exit int
	var err error
	if i < len(r.outs) {
		out = r.outs[i]
	}
	if i < len(r.exits) {
		exit = r.exits[i]
	}
	if i < len(r.errs) {
		err = r.errs[i]
	}
	return out, exit, err
}

func TestIdentifyUsesSmartctlExactArgumentsAndPreservesHealthExit(t *testing.T) {
	runner := &fakeResultRunner{outs: [][]byte{[]byte(`{"smart_status":{"passed":false}}`)}, exits: []int{8}}
	device := Device{Path: "/dev/sdb", KernelName: "sdb", AcquisitionCandidate: true}
	got, err := Identify(context.Background(), runner, device, "inventory-id", time.Now())
	if err != nil || len(got.Commands) != 1 || got.Commands[0].Status != "HEALTH_FAILED" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	want := resultCall{"smartctl", []string{"--json=c", "--all", "/dev/sdb"}}
	if !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("call=%+v", runner.calls[0])
	}
}

func TestIdentifyNVMeUsesTwoStructuredQueries(t *testing.T) {
	runner := &fakeResultRunner{outs: [][]byte{[]byte(`{"sn":"abc"}`), []byte(`{"critical_warning":0}`)}}
	device := Device{Path: "/dev/nvme1n1", KernelName: "nvme1n1", Transport: "nvme", AcquisitionCandidate: true}
	got, err := Identify(context.Background(), runner, device, "inventory-id", time.Now())
	if err != nil || len(got.Commands) != 2 || runner.calls[0].name != "nvme" || runner.calls[0].args[0] != "id-ctrl" || runner.calls[1].args[0] != "smart-log" {
		t.Fatalf("got=%+v calls=%+v err=%v", got, runner.calls, err)
	}
}

func TestIdentifyDistinguishesUnavailableAndInvalidOutput(t *testing.T) {
	runner := &fakeResultRunner{outs: [][]byte{nil}, errs: []error{errors.New("not found")}}
	got, err := Identify(context.Background(), runner, Device{Path: "/dev/sdb", AcquisitionCandidate: true}, "inventory-id", time.Now())
	if err != nil || got.Commands[0].Status != "UNAVAILABLE" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	runner = &fakeResultRunner{outs: [][]byte{[]byte("text")}}
	got, err = Identify(context.Background(), runner, Device{Path: "/dev/sdb", AcquisitionCandidate: true}, "inventory-id", time.Now())
	if err != nil || got.Commands[0].Status != "INVALID_OUTPUT" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}
