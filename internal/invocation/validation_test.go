package invocation

import (
	"strings"
	"testing"
	"time"
)

func TestContainerInvocationValidation(t *testing.T) {
	r := Record{Schema: Schema, ID: "77777777-7777-7777-8777-777777777777", WorkbookID: "88888888-8888-7888-8888-888888888888", Started: time.Unix(1, 0), Ended: time.Unix(2, 0), Executor: "container", Executable: "tool", Arguments: []string{"tool"}, Status: "complete", Stdout: "tool-output/a.stdout", Stderr: "tool-output/a.stderr", ContainerImage: "tool:1", ContainerDigest: "sha256:" + strings.Repeat("a", 64)}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	r.ContainerDigest = "latest"
	if err := r.Validate(); err == nil {
		t.Fatal("mutable container identity accepted")
	}
}

func TestInvocationValidationRejectsUnsafePaths(t *testing.T) {
	r := Record{Schema: Schema, ID: "77777777-7777-7777-8777-777777777777", WorkbookID: "88888888-8888-7888-8888-888888888888", Started: time.Unix(1, 0), Ended: time.Unix(2, 0), Executor: "native", Executable: "/bin/true", Arguments: []string{"true"}, Status: "complete", Stdout: "../escape", Stderr: "tool-output/a.stderr"}
	if err := r.Validate(); err == nil {
		t.Fatal("unsafe artifact path accepted")
	}
	r.Stdout = "tool-output/a.stdout"
	r.ID = "not-an-id"
	if err := r.Validate(); err == nil {
		t.Fatal("invalid invocation ID accepted")
	}
}
