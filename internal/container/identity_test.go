package container

import (
	"context"
	"testing"
)

type fakeRunner struct {
	output string
	err    error
}

func (f fakeRunner) Run(context.Context, string, ...string) (string, error) { return f.output, f.err }

func TestResolveRequiresDigest(t *testing.T) {
	id, err := Resolve(context.Background(), fakeRunner{output: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n"}, "registry.example/tool:1")
	if err != nil || id.Digest == "" {
		t.Fatalf("resolve: %+v %v", id, err)
	}
	if _, err := Resolve(context.Background(), fakeRunner{output: "latest"}, "tool:latest"); err == nil {
		t.Fatal("mutable tag accepted")
	}
	if _, err := Resolve(context.Background(), fakeRunner{output: "sha256:" + "g" + string(make([]byte, 63))}, "tool:bad"); err == nil {
		t.Fatal("non-hex digest accepted")
	}
}

func TestPodmanRunnerUsesArgumentVector(t *testing.T) {
	out, err := (PodmanRunner{}).Run(context.Background(), "printf", "%s", "ok")
	if err != nil || out != "ok" {
		t.Fatalf("runner: %q %v", out, err)
	}
}
