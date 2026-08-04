package container

import (
	"context"
	"testing"
)

type captureRunner struct {
	command string
	args    []string
}

func (r *captureRunner) Run(_ context.Context, command string, args ...string) (string, error) {
	r.command = command
	r.args = args
	return "ok", nil
}
func TestExecuteBuildsPodmanVector(t *testing.T) {
	r := &captureRunner{}
	out, err := Execute(context.Background(), r, Spec{Identity: Identity{Image: "tool", Digest: "sha256:" + "a" + string(make([]byte, 63))}, Arguments: []string{"tool"}})
	if err != nil || out != "ok" || r.command != "podman" || len(r.args) == 0 || r.args[0] != "run" {
		t.Fatalf("execute: %q %v %#v", out, err, r.args)
	}
}
