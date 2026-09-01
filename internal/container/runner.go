package container

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type PodmanRunner struct{}

func (PodmanRunner) Run(ctx context.Context, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(out))
		if detail != "" {
			return string(out), fmt.Errorf("%w: %s", err, detail)
		}
	}
	return string(out), err
}
