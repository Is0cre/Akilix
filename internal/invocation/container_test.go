package invocation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	containerpkg "github.com/pensuse/pensuse/internal/container"
)

func TestRunContainerRecordsDigestAndOutput(t *testing.T) {
	bin := t.TempDir()
	fake := filepath.Join(bin, "podman")
	script := "#!/bin/sh\nprintf container-out\nprintf container-err >&2\n"
	if err := os.WriteFile(fake, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	spec := containerpkg.Spec{Identity: containerpkg.Identity{Image: "local/tool", Digest: "sha256:" + "a" + string(make([]byte, 63))}, Arguments: []string{"tool"}}
	r, err := RunContainer(context.Background(), root, "wb", spec, func() time.Time { return time.Unix(10, 0) }, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Executor != "container" || r.ContainerDigest == "" {
		t.Fatalf("record: %+v", r)
	}
	out, _ := os.ReadFile(filepath.Join(root, r.Stdout))
	if string(out) != "container-out" {
		t.Fatalf("stdout: %q", out)
	}
	if _, err := List(root); err != nil {
		t.Fatal(err)
	}
}
