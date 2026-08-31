package acquire

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const InspectionSchema = "akilix.hardware-inspection.v1"

var lsblkColumns = "NAME,KNAME,PATH,TYPE,SIZE,RO,RM,TRAN,VENDOR,MODEL,SERIAL,WWN,FSTYPE,UUID,MOUNTPOINTS,PKNAME"

type Runner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok && len(exit.Stderr) != 0 {
			return nil, fmt.Errorf("%s: %s", name, strings.TrimSpace(string(exit.Stderr)))
		}
		return nil, err
	}
	return out, nil
}

type Inspection struct {
	Schema      string    `json:"schema"`
	GeneratedAt time.Time `json:"generated_at"`
	Passive     bool      `json:"passive"`
	Source      string    `json:"source"`
	Devices     []Device  `json:"devices"`
}

type Device struct {
	Name                 string      `json:"name"`
	KernelName           string      `json:"kernel_name"`
	Path                 string      `json:"path"`
	SizeBytes            uint64      `json:"size_bytes"`
	ReadOnly             bool        `json:"read_only"`
	Removable            bool        `json:"removable"`
	Transport            string      `json:"transport,omitempty"`
	Vendor               string      `json:"vendor,omitempty"`
	Model                string      `json:"model,omitempty"`
	Serial               string      `json:"serial,omitempty"`
	WWN                  string      `json:"wwn,omitempty"`
	Trusted              bool        `json:"trusted"`
	TrustID              string      `json:"trust_id,omitempty"`
	SystemDisk           bool        `json:"system_disk"`
	AcquisitionCandidate bool        `json:"acquisition_candidate"`
	Mounted              bool        `json:"mounted"`
	Mountpoints          []string    `json:"mountpoints"`
	Partitions           []Partition `json:"partitions"`
}

func ApplyTrust(report *Inspection, registry TrustRegistry) {
	for i := range report.Devices {
		entry, ok := registry.Match(report.Devices[i])
		report.Devices[i].Trusted = ok
		if ok {
			report.Devices[i].TrustID = entry.ID
		} else {
			report.Devices[i].TrustID = ""
		}
	}
}

type Partition struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	SizeBytes   uint64   `json:"size_bytes"`
	ReadOnly    bool     `json:"read_only"`
	Filesystem  string   `json:"filesystem,omitempty"`
	UUID        string   `json:"uuid,omitempty"`
	Mounted     bool     `json:"mounted"`
	Mountpoints []string `json:"mountpoints"`
}

type lsblkReport struct {
	Devices []lsblkDevice `json:"blockdevices"`
}

type lsblkDevice struct {
	Name        string        `json:"name"`
	KernelName  string        `json:"kname"`
	Path        string        `json:"path"`
	Type        string        `json:"type"`
	Size        uint64        `json:"size"`
	ReadOnly    bool          `json:"ro"`
	Removable   bool          `json:"rm"`
	Transport   string        `json:"tran"`
	Vendor      string        `json:"vendor"`
	Model       string        `json:"model"`
	Serial      string        `json:"serial"`
	WWN         string        `json:"wwn"`
	Filesystem  string        `json:"fstype"`
	UUID        string        `json:"uuid"`
	Mountpoints []interface{} `json:"mountpoints"`
	ParentName  string        `json:"pkname"`
	Children    []lsblkDevice `json:"children"`
}

// Inspect performs one passive, structured block-device inventory. It does not
// open device nodes, mount filesystems, or change kernel device state.
func Inspect(ctx context.Context, runner Runner, now time.Time) (Inspection, error) {
	args := []string{"--json", "--bytes", "--tree", "--output", lsblkColumns}
	out, err := runner.Run(ctx, "lsblk", args...)
	if err != nil {
		return Inspection{}, fmt.Errorf("inspect block devices: %w", err)
	}
	var raw lsblkReport
	if err := json.Unmarshal(out, &raw); err != nil {
		return Inspection{}, fmt.Errorf("decode lsblk JSON: %w", err)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	report := Inspection{Schema: InspectionSchema, GeneratedAt: now.UTC(), Passive: true, Source: "lsblk-json", Devices: []Device{}}
	for _, rawDevice := range raw.Devices {
		if rawDevice.Type != "disk" {
			continue
		}
		device := deviceFrom(rawDevice)
		device.SystemDisk = containsSystemMount(rawDevice)
		device.AcquisitionCandidate = !device.SystemDisk
		report.Devices = append(report.Devices, device)
	}
	sort.Slice(report.Devices, func(i, j int) bool { return report.Devices[i].Path < report.Devices[j].Path })
	return report, nil
}

func deviceFrom(raw lsblkDevice) Device {
	mounts := collectMounts(raw)
	device := Device{Name: clean(raw.Name), KernelName: clean(raw.KernelName), Path: clean(raw.Path), SizeBytes: raw.Size, ReadOnly: raw.ReadOnly, Removable: raw.Removable, Transport: clean(raw.Transport), Vendor: clean(raw.Vendor), Model: clean(raw.Model), Serial: clean(raw.Serial), WWN: clean(raw.WWN), Mounted: len(mounts) != 0, Mountpoints: mounts, Partitions: []Partition{}}
	appendPartitions(&device.Partitions, raw.Children)
	sort.Slice(device.Partitions, func(i, j int) bool { return device.Partitions[i].Path < device.Partitions[j].Path })
	return device
}

func appendPartitions(out *[]Partition, children []lsblkDevice) {
	for _, child := range children {
		if child.Type == "part" {
			mounts := mountpoints(child.Mountpoints)
			*out = append(*out, Partition{Name: clean(child.Name), Path: clean(child.Path), SizeBytes: child.Size, ReadOnly: child.ReadOnly, Filesystem: clean(child.Filesystem), UUID: clean(child.UUID), Mounted: len(mounts) != 0, Mountpoints: mounts})
		}
		appendPartitions(out, child.Children)
	}
}

func containsSystemMount(device lsblkDevice) bool {
	for _, mount := range mountpoints(device.Mountpoints) {
		if mount == "/" || mount == "/boot" || mount == "/boot/efi" {
			return true
		}
	}
	for _, child := range device.Children {
		if containsSystemMount(child) {
			return true
		}
	}
	return false
}

func collectMounts(device lsblkDevice) []string {
	set := map[string]bool{}
	var visit func(lsblkDevice)
	visit = func(item lsblkDevice) {
		for _, mount := range mountpoints(item.Mountpoints) {
			set[mount] = true
		}
		for _, child := range item.Children {
			visit(child)
		}
	}
	visit(device)
	result := make([]string, 0, len(set))
	for mount := range set {
		result = append(result, mount)
	}
	sort.Strings(result)
	return result
}

func mountpoints(values []interface{}) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if mount, ok := value.(string); ok && strings.TrimSpace(mount) != "" {
			result = append(result, strings.TrimSpace(mount))
		}
	}
	sort.Strings(result)
	return result
}

func clean(value string) string { return strings.TrimSpace(value) }
