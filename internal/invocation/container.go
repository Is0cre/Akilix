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

type containerManifest struct {
	Schema       string            `json:"schema"`
	InvocationID string            `json:"invocation_id"`
	Image        string            `json:"image"`
	Digest       string            `json:"digest"`
	Arguments    []string          `json:"arguments"`
	Network      string            `json:"network"`
	WritableRoot bool              `json:"writable_root"`
	Environment  map[string]string `json:"environment,omitempty"`
	Workdir      string            `json:"workdir,omitempty"`
	Status       string            `json:"status"`
}

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
	if syncErr := out.Sync(); syncErr != nil && runErr == nil {
		runErr = syncErr
	}
	if syncErr := errFile.Sync(); syncErr != nil && runErr == nil {
		runErr = syncErr
	}
	end := now().UTC()
	exitCode, status := 0, "complete"
	if runErr != nil {
		status, exitCode = "failed", 1
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}
	r := Record{Schema: Schema, ID: id, WorkbookID: workbookID, Started: start, Ended: end, Executor: "container", Executable: executable, Arguments: args, WorkingDirectory: spec.Workdir, Environment: spec.Environment, ExitCode: exitCode, Status: status, Stdout: filepath.ToSlash(filepath.Join("tool-output", id+".stdout")), Stderr: filepath.ToSlash(filepath.Join("tool-output", id+".stderr")), ScopeResult: options.ScopeResult, ScopeOverride: options.ScopeOverride, ContainerImage: spec.Identity.Image, ContainerDigest: spec.Identity.Digest}
	if err := r.Validate(); err != nil {
		return Record{}, err
	}
	containerData, err := json.MarshalIndent(containerManifest{Schema: "pensuse.container.v1", InvocationID: id, Image: spec.Identity.Image, Digest: spec.Identity.Digest, Arguments: spec.Arguments, Network: spec.Network, WritableRoot: spec.WritableRoot, Environment: spec.Environment, Workdir: spec.Workdir, Status: status}, "", "  ")
	if err != nil {
		return Record{}, err
	}
	containerData = append(containerData, '\n')
	if err := atomicWriteFile(filepath.Join(workbookRoot, ".pensuse", "containers", id+".json"), containerData); err != nil {
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

func atomicWriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(data)
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	return err
}
