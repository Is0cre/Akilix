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
	if !validDigest(digest) {
		return Identity{}, fmt.Errorf("container image %q did not resolve to an immutable sha256 digest", image)
	}
	return Identity{Image: image, Digest: digest}, nil
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	for _, c := range value[len("sha256:"):] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
