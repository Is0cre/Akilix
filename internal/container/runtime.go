package container

import (
	"context"
	"fmt"
	"strings"
)

// RuntimeStatus reports only local Podman state. Checking it never pulls an
// image, contacts a registry, or changes the operator's container storage.
type RuntimeStatus struct {
	Runtime            string `json:"runtime"`
	Available          bool   `json:"available"`
	Rootless           bool   `json:"rootless"`
	UserNamespaceReady bool   `json:"user_namespace_ready"`
}

func CheckRuntime(ctx context.Context, runner Runner) (RuntimeStatus, error) {
	status := RuntimeStatus{Runtime: "podman"}
	out, err := runner.Run(ctx, "podman", "info", "--format", "{{.Host.Security.Rootless}}")
	if err != nil {
		return status, fmt.Errorf("podman info: %w", err)
	}
	status.Available = true
	status.Rootless = strings.EqualFold(strings.TrimSpace(out), "true")
	if !status.Rootless {
		return status, fmt.Errorf("Podman is available but is not running rootless")
	}
	if _, err := runner.Run(ctx, "podman", "unshare", "true"); err != nil {
		return status, fmt.Errorf("rootless Podman user namespace: %w", err)
	}
	status.UserNamespaceReady = true
	return status, nil
}
