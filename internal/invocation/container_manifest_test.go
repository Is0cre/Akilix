package invocation

import (
	"testing"

	containerpkg "github.com/pensuse/pensuse/internal/container"
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
