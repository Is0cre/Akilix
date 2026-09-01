package profile

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

const PreflightSchema = "akilix.profile-preflight.v1"

type PreflightCheck struct {
	Name    string   `json:"name"`
	Ready   bool     `json:"ready"`
	Detail  string   `json:"detail"`
	Command []string `json:"command,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type Preflight struct {
	Schema            string           `json:"schema"`
	ProfileID         string           `json:"profile_id"`
	Ready             bool             `json:"ready"`
	ApplySupported    bool             `json:"apply_supported"`
	RequiresPrivilege bool             `json:"requires_privilege"`
	Checks            []PreflightCheck `json:"checks"`
}

type HostInspector interface {
	Run(context.Context, string, ...string) ([]byte, error)
	LookPath(string) (string, error)
	EUID() int
}

type ExecInspector struct{}

func (ExecInspector) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
func (ExecInspector) LookPath(name string) (string, error) { return exec.LookPath(name) }
func (ExecInspector) EUID() int                            { return os.Geteuid() }

// PreflightHost performs local, read-only readiness checks. It does not create
// a snapshot or invoke zypper, rpm transactions, repository refreshes, or OCI
// operations.
func PreflightHost(ctx context.Context, manifest Manifest, host HostInspector) (Preflight, error) {
	if err := manifest.Validate(); err != nil {
		return Preflight{}, err
	}
	result := Preflight{
		Schema: PreflightSchema, ProfileID: manifest.ID, Ready: true,
		ApplySupported: false, RequiresPrivilege: host.EUID() != 0,
		Checks: []PreflightCheck{},
	}
	for _, tool := range []string{"btrfs", "findmnt", "rpm", "snapper", "zypper"} {
		path, err := host.LookPath(tool)
		check := PreflightCheck{Name: "tool:" + tool, Ready: err == nil, Detail: path}
		if err != nil {
			check.Detail = "required host tool unavailable"
			check.Error = boundedDetail(nil, err)
			result.Ready = false
		}
		result.Checks = append(result.Checks, check)
	}
	out, err := host.Run(ctx, "findmnt", "-n", "-o", "FSTYPE", "/")
	filesystem := strings.TrimSpace(string(out))
	btrfsReady := err == nil && filesystem == "btrfs"
	filesystemCheck := PreflightCheck{Name: "root-filesystem", Ready: btrfsReady, Detail: filesystem, Command: []string{"findmnt", "-n", "-o", "FSTYPE", "/"}}
	if err != nil {
		filesystemCheck.Detail = "unable to inspect root filesystem"
		filesystemCheck.Error = boundedDetail(out, err)
	} else if filesystem == "" {
		filesystemCheck.Detail = "filesystem type unavailable"
	}
	result.Checks = append(result.Checks, filesystemCheck)
	if !btrfsReady {
		result.Ready = false
	}
	out, err = host.Run(ctx, "snapper", "--no-dbus", "list-configs", "--columns", "config,subvolume")
	snapperReady := err == nil && hasRootSnapperConfig(string(out))
	snapperCheck := PreflightCheck{Name: "snapper-root-config", Ready: snapperReady, Detail: "root subvolume configured", Command: []string{"snapper", "--no-dbus", "list-configs", "--columns", "config,subvolume"}}
	if err != nil {
		snapperCheck.Detail = "unable to inspect Snapper configurations"
		snapperCheck.Error = boundedDetail(out, err)
	} else if !snapperReady {
		snapperCheck.Detail = "root subvolume not configured"
	}
	result.Checks = append(result.Checks, snapperCheck)
	if !snapperReady {
		result.Ready = false
	}
	statusReady := manifest.Status == "foundation"
	result.Checks = append(result.Checks, PreflightCheck{Name: "profile-maturity", Ready: statusReady, Detail: manifest.Status})
	if !statusReady {
		result.Ready = false
	}
	return result, nil
}

func hasRootSnapperConfig(output string) bool {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[len(fields)-1] == "/" {
			return true
		}
	}
	return false
}
