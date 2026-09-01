package profile

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

const PreflightSchema = "akilix.profile-preflight.v1"

type PreflightCheck struct {
	Name   string `json:"name"`
	Ready  bool   `json:"ready"`
	Detail string `json:"detail"`
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
	for _, tool := range []string{"btrfs", "rpm", "snapper", "zypper"} {
		path, err := host.LookPath(tool)
		result.Checks = append(result.Checks, PreflightCheck{Name: "tool:" + tool, Ready: err == nil, Detail: path})
		if err != nil {
			result.Ready = false
		}
	}
	out, err := host.Run(ctx, "findmnt", "-n", "-o", "FSTYPE", "/")
	btrfsReady := err == nil && strings.TrimSpace(string(out)) == "btrfs"
	result.Checks = append(result.Checks, PreflightCheck{Name: "root-filesystem", Ready: btrfsReady, Detail: strings.TrimSpace(string(out))})
	if !btrfsReady {
		result.Ready = false
	}
	out, err = host.Run(ctx, "snapper", "--no-dbus", "list-configs", "--columns", "config,subvolume")
	snapperReady := err == nil && hasRootSnapperConfig(string(out))
	result.Checks = append(result.Checks, PreflightCheck{Name: "snapper-root-config", Ready: snapperReady, Detail: "root subvolume configured"})
	if !snapperReady {
		result.Checks[len(result.Checks)-1].Detail = "root subvolume not configured"
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
