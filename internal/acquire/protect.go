package acquire

import (
	"context"
	"fmt"
	"strings"
)

type ProtectionResult struct {
	Device          Device `json:"device"`
	AlreadyReadOnly bool   `json:"already_read_only"`
	KernelReadOnly  bool   `json:"kernel_read_only"`
}

func Candidate(report Inspection, path string) (Device, error) {
	for _, device := range report.Devices {
		if device.Path != path {
			continue
		}
		if device.SystemDisk || !device.AcquisitionCandidate {
			return Device{}, fmt.Errorf("refusing system or non-candidate device %q", path)
		}
		if device.Mounted {
			return Device{}, fmt.Errorf("refusing mounted device %q", path)
		}
		return device, nil
	}
	return Device{}, fmt.Errorf("device %q is not an inspected whole disk", path)
}

// SetReadOnly changes only the kernel block-device read-only flag and verifies
// the resulting state. It never invokes sudo or unmounts filesystems.
func SetReadOnly(ctx context.Context, runner Runner, device Device) (ProtectionResult, error) {
	if device.Path == "" || device.SystemDisk || !device.AcquisitionCandidate || device.Mounted {
		return ProtectionResult{}, fmt.Errorf("invalid protection candidate")
	}
	result := ProtectionResult{Device: device, AlreadyReadOnly: device.ReadOnly}
	if !device.ReadOnly {
		if _, err := runner.Run(ctx, "blockdev", "--setro", device.Path); err != nil {
			return result, fmt.Errorf("set kernel read-only flag: %w", err)
		}
	}
	out, err := runner.Run(ctx, "blockdev", "--getro", device.Path)
	if err != nil {
		return result, fmt.Errorf("verify kernel read-only flag: %w", err)
	}
	result.KernelReadOnly = strings.TrimSpace(string(out)) == "1"
	if !result.KernelReadOnly {
		return result, fmt.Errorf("kernel read-only verification failed for %q", device.Path)
	}
	return result, nil
}
