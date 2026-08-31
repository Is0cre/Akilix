package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/pensuse/pensuse/internal/version"
	"github.com/pensuse/pensuse/internal/workbook"
)

func TestVersionJSONShape(t *testing.T) {
	b, err := json.Marshal(version.Current())
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"name", "version", "base", "architecture"} {
		if got[key] == "" {
			t.Fatalf("missing %s", key)
		}
	}
}

func TestRunVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"version"}, &out, &errOut); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut.String())
	}
	if out.String() == "" || errOut.Len() != 0 {
		t.Fatalf("unexpected output: stdout=%q stderr=%q", out.String(), errOut.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"status"}, &out, &errOut); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestContainerRunRejectsUnknownAndDuplicateOptions(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"container", "run", "case-1", "image", "--bogus", "--", "true"}, &out, &errOut); code != 2 {
		t.Fatalf("unknown container option exit code = %d, want 2", code)
	}
	errOut.Reset()
	if code := run([]string{"container", "run", "case-1", "image", "--env", "A=1", "--env", "A=2", "--", "true"}, &out, &errOut); code != 2 {
		t.Fatalf("duplicate container env exit code = %d, want 2", code)
	}
}

func TestScopeClosedWorkbookCannotMutateAndJSONListIsJSON(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PENSUSE_WORKBOOK_ROOT", root)
	if _, err := workbook.Create(root, "case-1", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"scope", "add", "case-1", "Example.COM"}, &out, &errOut); code != 0 {
		t.Fatalf("scope add: %d %s", code, errOut.String())
	}
	if _, err := workbook.SetStatus(root, "case-1", "closed"); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errOut.Reset()
	if code := run([]string{"scope", "add", "case-1", "other.example"}, &out, &errOut); code == 0 {
		t.Fatal("closed scope mutation accepted")
	}
	out.Reset()
	errOut.Reset()
	if code := run([]string{"scope", "list", "case-1", "--json"}, &out, &errOut); code != 0 {
		t.Fatalf("scope JSON list: %d %s", code, errOut.String())
	}
	var got map[string][]string
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("scope list was not JSON: %v", err)
	}
	if len(got["include"]) != 1 || got["include"][0] != "example.com" {
		t.Fatalf("unexpected scope: %+v", got)
	}
}

func TestLoggingStatusIsExplicitAndJSONReadable(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PENSUSE_WORKBOOK_ROOT", root)
	if _, err := workbook.Create(root, "case-1", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"logging", "status", "case-1"}, &out, &errOut); code != 0 {
		t.Fatalf("logging status: %d %s", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Terminal recording       disabled")) {
		t.Fatalf("logging status omitted disabled sensitive feature: %q", out.String())
	}
	out.Reset()
	errOut.Reset()
	if code := run([]string{"logging", "status", "case-1", "--json"}, &out, &errOut); code != 0 {
		t.Fatalf("logging JSON status: %d %s", code, errOut.String())
	}
	var policy map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &policy); err != nil {
		t.Fatalf("logging status was not JSON: %v", err)
	}
	if policy["terminal_recording"] != false || policy["command_metadata"] != true {
		t.Fatalf("unexpected logging policy: %+v", policy)
	}
}
