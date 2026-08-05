package container

import "testing"

func TestSpecRequiresReadOnlyOriginals(t *testing.T) {
	s := Spec{Identity: Identity{Digest: "sha256:" + "a" + string(make([]byte, 63))}, Arguments: []string{"tool"}, Mounts: []Mount{{Source: "/evidence", Destination: "/input", OriginalEvidence: true}}}
	if err := s.Validate(); err == nil {
		t.Fatal("writable original accepted")
	}
	s.Mounts[0].ReadOnly = true
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	args, err := s.Args()
	if err != nil || len(args) < 15 || args[0] != "run" || args[3] != "--network=none" || args[4] != "--userns=keep-id" || args[5] != "--pid=private" || args[6] != "--ipc=private" || args[7] != "--uts=private" || args[8] != "--security-opt=no-new-privileges" || args[9] != "--cap-drop=ALL" || args[10] != "--read-only" {
		t.Fatalf("args: %#v %v", args, err)
	}
}

func TestSpecEnvironmentAndWorkdirValidation(t *testing.T) {
	base := Spec{Identity: Identity{Image: "tool", Digest: "sha256:" + "a" + string(make([]byte, 63))}, Arguments: []string{"tool"}, Environment: map[string]string{"Z": "2", "A": "1"}, Workdir: "/work"}
	args, err := base.Args()
	if err != nil {
		t.Fatal(err)
	}
	if args[11] != "--env" || args[12] != "A=1" || args[13] != "--env" || args[14] != "Z=2" {
		t.Fatalf("environment ordering: %#v", args)
	}
	base.Workdir = "relative"
	if _, err := base.Args(); err == nil {
		t.Fatal("relative workdir accepted")
	}
}

func TestSpecRejectsUnsafeMountAndEnvironmentValues(t *testing.T) {
	base := Spec{Identity: Identity{Image: "tool", Digest: "sha256:" + "a" + string(make([]byte, 63))}, Arguments: []string{"tool"}}
	base.Mounts = []Mount{{Source: "/tmp/../evidence", Destination: "/input"}}
	if _, err := base.Args(); err == nil {
		t.Fatal("unclean mount path accepted")
	}
	base.Mounts = nil
	base.Environment = map[string]string{"SAFE": "bad\x00value"}
	if _, err := base.Args(); err == nil {
		t.Fatal("NUL environment value accepted")
	}
}
