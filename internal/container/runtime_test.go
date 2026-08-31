package container

import (
	"context"
	"errors"
	"testing"
)

type runtimeRunner struct {
	calls    int
	rootless string
	err      error
}

func (r *runtimeRunner) Run(_ context.Context, command string, args ...string) (string, error) {
	r.calls++
	if command != "podman" {
		return "", errors.New("unexpected runtime")
	}
	if r.calls == 1 {
		return r.rootless, r.err
	}
	return "", r.err
}

func TestCheckRuntimeRequiresRootlessUserNamespace(t *testing.T) {
	runner := &runtimeRunner{rootless: "true\n"}
	status, err := CheckRuntime(context.Background(), runner)
	if err != nil || !status.Available || !status.Rootless || !status.UserNamespaceReady || runner.calls != 2 {
		t.Fatalf("status=%+v calls=%d err=%v", status, runner.calls, err)
	}

	runner = &runtimeRunner{rootless: "false\n"}
	status, err = CheckRuntime(context.Background(), runner)
	if err == nil || !status.Available || status.Rootless || runner.calls != 1 {
		t.Fatalf("rootful runtime accepted: status=%+v calls=%d err=%v", status, runner.calls, err)
	}
}
