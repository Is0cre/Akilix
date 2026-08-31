package logpolicy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultRoundTripAndTransparentOptIn(t *testing.T) {
	p := Default()
	if !p.CommandMetadata || !p.EvidenceHashing || !p.StdoutCapture || p.PacketMetadata || p.TerminalRecording {
		t.Fatalf("unsafe default logging policy: %+v", p)
	}
	rendered, err := Render(p)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "logging.yaml"), []byte(rendered), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(root)
	if err != nil || loaded != p {
		t.Fatalf("round trip: %+v %v", loaded, err)
	}
}

func TestLoadRejectsMissingAndUnknownKeys(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "logging.yaml"), []byte("schema: pensuse.logging.v1\nterminal_recording: false\nsecret_capture: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("accepted incomplete policy with unknown logging feature")
	}
}
