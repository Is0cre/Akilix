package invocation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	Policy       containerPolicy   `json:"policy"`
	Mounts       []containerMount  `json:"mounts,omitempty"`
	Generated    []string          `json:"generated_files,omitempty"`
	Status       string            `json:"status"`
}

type containerPolicy struct {
	Pull         string `json:"pull"`
	UserNS       string `json:"user_namespace"`
	PIDNS        string `json:"pid_namespace"`
	IPCNS        string `json:"ipc_namespace"`
	UTSNS        string `json:"uts_namespace"`
	NoNewPrivs   bool   `json:"no_new_privileges"`
	Capabilities string `json:"capabilities"`
	RootReadOnly bool   `json:"root_read_only"`
}

type containerMount struct {
	Source           string `json:"source"`
	Destination      string `json:"destination"`
	ReadOnly         bool   `json:"read_only"`
	OriginalEvidence bool   `json:"original_evidence"`
}

func RunContainer(ctx context.Context, workbookRoot, workbookID string, spec containerpkg.Spec, now func() time.Time, options Options) (Record, error) {
	executable, err := exec.LookPath("podman")
	if err != nil {
		return Record{}, err
	}
	start := now().UTC()
	id, err := uuid(start)
	if err != nil {
		return Record{}, err
	}
	var invocationOutputDir string
	if spec.InvocationOutput {
		invocationOutputDir = filepath.Join(workbookRoot, "artifacts", "derived", id)
		if err := os.Mkdir(invocationOutputDir, 0700); err != nil {
			return Record{}, fmt.Errorf("create invocation output directory: %w", err)
		}
		spec.Mounts = append(spec.Mounts, containerpkg.Mount{Source: invocationOutputDir, Destination: "/workbook/output"})
	}
	args, err := spec.Args()
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
	before, err := snapshotToolOutput(outDir)
	if err != nil {
		return Record{}, err
	}
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
	generated, generatedErr := generatedToolOutput(outDir, before)
	if generatedErr == nil && invocationOutputDir != "" {
		outputFiles, outputErr := generatedInvocationOutput(workbookRoot, invocationOutputDir)
		if outputErr != nil {
			generatedErr = outputErr
		} else {
			generated = append(generated, outputFiles...)
		}
	}
	if generatedErr != nil && runErr == nil {
		runErr = generatedErr
	}
	exitCode, status := 0, "complete"
	if runErr != nil {
		status, exitCode = "failed", 1
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}
	r := Record{Schema: Schema, ID: id, WorkbookID: workbookID, Started: start, Ended: end, Executor: "container", Executable: executable, Arguments: args, WorkingDirectory: spec.Workdir, Environment: spec.Environment, GeneratedFiles: generated, ExitCode: exitCode, Status: status, Stdout: filepath.ToSlash(filepath.Join("tool-output", id+".stdout")), Stderr: filepath.ToSlash(filepath.Join("tool-output", id+".stderr")), ScopeResult: options.ScopeResult, ScopeTarget: options.ScopeTarget, ScopeOverride: options.ScopeOverride, ContainerImage: spec.Identity.Image, ContainerDigest: spec.Identity.Digest}
	if err := r.Validate(); err != nil {
		return Record{}, err
	}
	containerData, err := json.MarshalIndent(buildContainerManifest(id, spec, generated, status), "", "  ")
	if err != nil {
		return Record{}, err
	}
	containerData = append(containerData, '\n')
	if err := atomicWriteFile(filepath.Join(workbookRoot, ".pensuse", "containers", id+".json"), containerData); err != nil {
		return Record{}, err
	}
	if err := appendRecord(workbookRoot, r); err != nil {
		return Record{}, err
	}
	if runErr != nil {
		return r, fmt.Errorf("container command failed: %w", runErr)
	}
	return r, nil
}

func generatedInvocationOutput(workbookRoot, outputDir string) ([]string, error) {
	var generated []string
	err := filepath.WalkDir(outputDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == outputDir || entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("container output contains symlink %q", path)
		}
		relative, err := filepath.Rel(workbookRoot, path)
		if err != nil {
			return err
		}
		generated = append(generated, filepath.ToSlash(relative))
		return nil
	})
	sort.Strings(generated)
	return generated, err
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
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err == nil {
		err = syncDirectory(filepath.Dir(path))
	}
	return err
}

func buildContainerManifest(id string, spec containerpkg.Spec, generated []string, status string) containerManifest {
	mounts := make([]containerMount, 0, len(spec.Mounts))
	for _, mount := range spec.Mounts {
		mounts = append(mounts, containerMount{Source: mount.Source, Destination: mount.Destination, ReadOnly: mount.ReadOnly, OriginalEvidence: mount.OriginalEvidence})
	}
	return containerManifest{
		Schema: "pensuse.container.v1", InvocationID: id, Image: spec.Identity.Image, Digest: spec.Identity.Digest,
		Arguments: spec.Arguments, Network: effectiveNetwork(spec.Network), WritableRoot: spec.WritableRoot,
		Environment: spec.Environment, Workdir: spec.Workdir, Mounts: mounts, Generated: generated, Status: status,
		Policy: containerPolicy{Pull: "never", UserNS: "keep-id", PIDNS: "private", IPCNS: "private", UTSNS: "private", NoNewPrivs: true, Capabilities: "drop-all", RootReadOnly: !spec.WritableRoot},
	}
}

func effectiveNetwork(network string) string {
	if network == "" {
		return "none"
	}
	return network
}

func syncDirectory(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
