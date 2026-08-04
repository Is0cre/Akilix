package container

import (
	"context"
	"fmt"
	"strings"
)

type Runner interface {
	Run(context.Context, string, ...string) (string, error)
}

type Identity struct {
	Image  string `json:"image"`
	Digest string `json:"digest"`
}

func Resolve(ctx context.Context, runner Runner, image string) (Identity, error) {
	if strings.TrimSpace(image) == "" {
		return Identity{}, fmt.Errorf("container image is required")
	}
	out, err := runner.Run(ctx, "podman", "image", "inspect", "--format", "{{.Digest}}", image)
	if err != nil {
		return Identity{}, err
	}
	digest := strings.TrimSpace(out)
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != len("sha256:")+64 {
		return Identity{}, fmt.Errorf("container image %q did not resolve to an immutable sha256 digest", image)
	}
	return Identity{Image: image, Digest: digest}, nil
}
