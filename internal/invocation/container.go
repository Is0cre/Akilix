package invocation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	containerpkg "github.com/pensuse/pensuse/internal/container"
)

func RunContainer(ctx context.Context, workbookRoot, workbookID string, spec containerpkg.Spec, now func() time.Time, options Options) (Record, error) {
	args, err := spec.Args()
	if err != nil {
		return Record{}, err
	}
	executable, err := exec.LookPath("podman")
	if err != nil {
		return Record{}, err
	}
	start := now().UTC()
	id, err := uuid(start)
	if err != nil {
		return Record{}, err
	}
	outDir := filepath.Join(workbookRoot, "tool-output")
	if err := os.MkdirAll(outDir, 0700); err != nil {
		return Record{}, err
	}
	outPath, errPath := filepath.Join(outDir, id+".stdout"), filepath.Join(outDir, id+".stderr")
	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return Record{}, err
	}
	defer out.Close()
	errFile, err := os.OpenFile(errPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return Record{}, err
	}
	defer errFile.Close()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdout, cmd.Stderr = out, errFile
	runErr := cmd.Run()
	end := now().UTC()
	exitCode, status := 0, "complete"
	if runErr != nil {
		status, exitCode = "failed", 1
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}
	r := Record{Schema: Schema, ID: id, WorkbookID: workbookID, Started: start, Ended: end, Executor: "container", Executable: executable, Arguments: args, ExitCode: exitCode, Status: status, Stdout: filepath.ToSlash(filepath.Join("tool-output", id+".stdout")), Stderr: filepath.ToSlash(filepath.Join("tool-output", id+".stderr")), ScopeResult: options.ScopeResult, ScopeOverride: options.ScopeOverride, ContainerImage: spec.Identity.Image, ContainerDigest: spec.Identity.Digest}
	if err := r.Validate(); err != nil {
		return Record{}, err
	}
	b, err := json.Marshal(r)
	if err != nil {
		return Record{}, err
	}
	b = append(b, '\n')
	manifest := filepath.Join(workbookRoot, ".pensuse", "manifest.jsonl")
	if err := os.MkdirAll(filepath.Dir(manifest), 0700); err != nil {
		return Record{}, err
	}
	f, err := os.OpenFile(manifest, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return Record{}, err
	}
	_, writeErr := f.Write(b)
	syncErr := f.Sync()
	closeErr := f.Close()
	if writeErr != nil {
		return Record{}, writeErr
	}
	if syncErr != nil {
		return Record{}, syncErr
	}
	if closeErr != nil {
		return Record{}, closeErr
	}
	if runErr != nil {
		return r, fmt.Errorf("container command failed: %w", runErr)
	}
	return r, nil
}
