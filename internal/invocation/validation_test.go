package invocation

import (
	"testing"
	"time"
)

func TestContainerInvocationValidation(t *testing.T) {
	r := Record{Schema: Schema, ID: "i", WorkbookID: "w", Started: time.Unix(1, 0), Ended: time.Unix(2, 0), Executor: "container", Arguments: []string{"tool"}, Status: "complete", ContainerImage: "tool:1", ContainerDigest: "sha256:" + "a" + string(make([]byte, 63))}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	r.ContainerDigest = "latest"
	if err := r.Validate(); err == nil {
		t.Fatal("mutable container identity accepted")
	}
}
