package invocation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const Schema = "pensuse.invocation.v1"

type Record struct {
	Schema          string    `json:"schema"`
	ID              string    `json:"invocation_id"`
	WorkbookID      string    `json:"workbook_id"`
	Started         time.Time `json:"started"`
	Ended           time.Time `json:"ended"`
	Executor        string    `json:"executor"`
	Executable      string    `json:"executable"`
	Arguments       []string  `json:"arguments"`
	ExitCode        int       `json:"exit_code"`
	Status          string    `json:"status"`
	Stdout          string    `json:"stdout_artifact"`
	Stderr          string    `json:"stderr_artifact"`
	ScopeResult     string    `json:"scope_result,omitempty"`
	ScopeOverride   bool      `json:"scope_override,omitempty"`
	ContainerImage  string    `json:"container_image,omitempty"`
	ContainerDigest string    `json:"container_digest,omitempty"`
}

type Options struct {
	ScopeResult   string
	ScopeOverride bool
}

func (r Record) Validate() error {
	if r.Schema != Schema || r.ID == "" || r.WorkbookID == "" || len(r.Arguments) == 0 {
		return fmt.Errorf("invalid invocation record")
	}
	if r.Ended.Before(r.Started) {
		return fmt.Errorf("invocation ended before it started")
	}
	if r.Executor == "container" && (r.ContainerImage == "" || !strings.HasPrefix(r.ContainerDigest, "sha256:") || len(r.ContainerDigest) != len("sha256:")+64) {
		return fmt.Errorf("container invocation requires immutable image digest")
	}
	return nil
}

func List(workbookRoot string) ([]Record, error) {
	b, err := os.ReadFile(filepath.Join(workbookRoot, ".pensuse", "manifest.jsonl"))
	if os.IsNotExist(err) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Record
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line == "" {
			continue
		}
		var r Record
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			return nil, err
		}
		if err := r.Validate(); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func Run(ctx context.Context, workbookRoot, workbookID string, args []string, now func() time.Time) (Record, error) {
	return RunWithOptions(ctx, workbookRoot, workbookID, args, now, Options{})
}

func RunWithOptions(ctx context.Context, workbookRoot, workbookID string, args []string, now func() time.Time, options Options) (Record, error) {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return Record{}, fmt.Errorf("command is required")
	}
	executable, err := exec.LookPath(args[0])
	if err != nil {
		return Record{}, err
	}
	start := now().UTC()
	id, err := uuid(start)
	if err != nil {
		return Record{}, err
	}
	outputDir := filepath.Join(workbookRoot, "tool-output")
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return Record{}, err
	}
	outPath := filepath.Join(outputDir, id+".stdout")
	errPath := filepath.Join(outputDir, id+".stderr")
	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return Record{}, err
	}
	defer out.Close()
	errFile, err := os.OpenFile(errPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		out.Close()
		return Record{}, err
	}
	defer errFile.Close()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = out
	cmd.Stderr = errFile
	runErr := cmd.Run()
	if syncErr := out.Sync(); syncErr != nil && runErr == nil {
		runErr = syncErr
	}
	if syncErr := errFile.Sync(); syncErr != nil && runErr == nil {
		runErr = syncErr
	}
	end := now().UTC()
	exitCode := 0
	status := "complete"
	if runErr != nil {
		status = "failed"
		exitCode = 1
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}
	r := Record{Schema: Schema, ID: id, WorkbookID: workbookID, Started: start, Ended: end, Executor: "native", Executable: executable, Arguments: append([]string(nil), args...), ExitCode: exitCode, Status: status, Stdout: filepath.ToSlash(filepath.Join("tool-output", id+".stdout")), Stderr: filepath.ToSlash(filepath.Join("tool-output", id+".stderr")), ScopeResult: options.ScopeResult, ScopeOverride: options.ScopeOverride}
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
		return r, fmt.Errorf("command failed: %w", runErr)
	}
	return r, nil
}

func uuid(t time.Time) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[6:]); err != nil {
		return "", err
	}
	ms := uint64(t.UnixMilli())
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	b[6] = (b[6] & 15) | 112
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:])), nil
}
