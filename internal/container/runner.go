package container

import (
	"context"
	"os/exec"
)

type PodmanRunner struct{}

func (PodmanRunner) Run(ctx context.Context, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	out, err := cmd.Output()
	return string(out), err
}
