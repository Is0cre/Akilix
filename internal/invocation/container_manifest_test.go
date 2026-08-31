package invocation

import (
	"os"
	"path/filepath"
	"testing"

	containerpkg "github.com/Is0cre/Akilix/internal/container"
)

func TestBuildContainerManifestCapturesEffectivePolicy(t *testing.T) {
	spec := containerpkg.Spec{
		Identity:  containerpkg.Identity{Image: "example/tool", Digest: "sha256:" + "a" + string(make([]byte, 63))},
		Arguments: []string{"tool"},
		Mounts:    []containerpkg.Mount{{Source: "/evidence", Destination: "/input", ReadOnly: true, OriginalEvidence: true}},
	}
	manifest := buildContainerManifest("id", spec, []string{"tool-output/result.txt"}, "complete")
	if manifest.Network != "none" || manifest.Policy.Pull != "never" || manifest.Policy.UserNS != "keep-id" || !manifest.Policy.NoNewPrivs || manifest.Policy.Capabilities != "drop-all" || !manifest.Policy.RootReadOnly {
		t.Fatalf("incomplete policy manifest: %+v", manifest)
	}
	if len(manifest.Mounts) != 1 || !manifest.Mounts[0].ReadOnly || !manifest.Mounts[0].OriginalEvidence {
		t.Fatalf("mount policy missing: %+v", manifest.Mounts)
	}
}

func TestGeneratedInvocationOutputRejectsSymlinks(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "artifacts", "derived", "inv")
	if err := os.MkdirAll(out, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "result.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	files, err := generatedInvocationOutput(root, out)
	if err != nil || len(files) != 1 || files[0] != "artifacts/derived/inv/result.json" {
		t.Fatalf("files=%v err=%v", files, err)
	}
	if err := os.Symlink("result.json", filepath.Join(out, "alias")); err != nil {
		t.Fatal(err)
	}
	if _, err := generatedInvocationOutput(root, out); err == nil {
		t.Fatal("symlinked output accepted")
	}
}
