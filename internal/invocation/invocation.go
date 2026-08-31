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
	"sort"
	"strings"
	"time"
)

const Schema = "pensuse.invocation.v1"

type Record struct {
	Schema           string            `json:"schema"`
	ID               string            `json:"invocation_id"`
	WorkbookID       string            `json:"workbook_id"`
	Started          time.Time         `json:"started"`
	Ended            time.Time         `json:"ended"`
	Executor         string            `json:"executor"`
	Executable       string            `json:"executable"`
	Arguments        []string          `json:"arguments"`
	WorkingDirectory string            `json:"working_directory,omitempty"`
	Environment      map[string]string `json:"environment,omitempty"`
	GeneratedFiles   []string          `json:"generated_files,omitempty"`
	ExitCode         int               `json:"exit_code"`
	Status           string            `json:"status"`
	Stdout           string            `json:"stdout_artifact"`
	Stderr           string            `json:"stderr_artifact"`
	ScopeResult      string            `json:"scope_result,omitempty"`
	ScopeTarget      string            `json:"scope_target,omitempty"`
	ScopeOverride    bool              `json:"scope_override,omitempty"`
	ContainerImage   string            `json:"container_image,omitempty"`
	ContainerDigest  string            `json:"container_digest,omitempty"`
}

type Options struct {
	ScopeResult   string
	ScopeTarget   string
	ScopeOverride bool
}

func (r Record) Validate() error {
	if r.Schema != Schema || !validID(r.ID) || !validID(r.WorkbookID) || len(r.Arguments) == 0 || strings.TrimSpace(r.Arguments[0]) == "" || r.Executable == "" || r.Started.IsZero() || r.Ended.IsZero() {
		return fmt.Errorf("invalid invocation record")
	}
	if r.Ended.Before(r.Started) || r.ExitCode < -1 {
		return fmt.Errorf("invalid invocation timing or exit code")
	}
	if r.Status != "complete" && r.Status != "failed" || r.Executor != "native" && r.Executor != "container" {
		return fmt.Errorf("invalid invocation status")
	}
	if !validArtifactPath(r.Stdout) || !validArtifactPath(r.Stderr) {
		return fmt.Errorf("invalid invocation artifact path")
	}
	if r.WorkingDirectory != "" && !filepath.IsAbs(r.WorkingDirectory) {
		return fmt.Errorf("working directory must be absolute")
	}
	for key, value := range r.Environment {
		if key == "" || strings.ContainsAny(key, "=\x00") || strings.ContainsRune(value, '\x00') {
			return fmt.Errorf("invalid invocation environment")
		}
	}
	for _, file := range r.GeneratedFiles {
		if !validArtifactPath(file) {
			return fmt.Errorf("invalid generated artifact path")
		}
	}
	if r.ScopeResult != "" && r.ScopeResult != "ALLOW" && r.ScopeResult != "DENY" && r.ScopeResult != "UNKNOWN" {
		return fmt.Errorf("invalid invocation scope result")
	}
	if r.ScopeResult != "" && r.ScopeTarget == "" || r.ScopeOverride && (r.ScopeResult != "DENY" || r.ScopeTarget == "") {
		return fmt.Errorf("invalid invocation scope provenance")
	}
	if r.Executor == "container" && (r.ContainerImage == "" || !validDigest(r.ContainerDigest)) {
		return fmt.Errorf("container invocation requires immutable image digest")
	}
	return nil
}

func validDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validHex(value[len("sha256:"):], 64)
}

func validHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, c := range value {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
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
	before, err := snapshotToolOutput(outputDir)
	if err != nil {
		return Record{}, err
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	executionEnvironment := safeEnvironment()
	cmd.Env = environmentList(executionEnvironment)
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
	generated, generatedErr := generatedToolOutput(outputDir, before)
	if generatedErr != nil && runErr == nil {
		runErr = generatedErr
	}
	exitCode := 0
	status := "complete"
	if runErr != nil {
		status = "failed"
		exitCode = 1
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}
	workingDirectory, _ := os.Getwd()
	r := Record{Schema: Schema, ID: id, WorkbookID: workbookID, Started: start, Ended: end, Executor: "native", Executable: executable, Arguments: append([]string(nil), args...), WorkingDirectory: workingDirectory, Environment: executionEnvironment, GeneratedFiles: generated, ExitCode: exitCode, Status: status, Stdout: filepath.ToSlash(filepath.Join("tool-output", id+".stdout")), Stderr: filepath.ToSlash(filepath.Join("tool-output", id+".stderr")), ScopeResult: options.ScopeResult, ScopeTarget: options.ScopeTarget, ScopeOverride: options.ScopeOverride}
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

func validID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || value[14] != '7' {
		return false
	}
	variant := value[19]
	if !((variant >= '8' && variant <= '9') || (variant >= 'a' && variant <= 'b') || (variant >= 'A' && variant <= 'B')) {
		return false
	}
	for i, c := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func validArtifactPath(value string) bool {
	slash := filepath.ToSlash(value)
	return value != "" && !filepath.IsAbs(value) && strings.HasPrefix(slash, "tool-output/") && !strings.Contains(slash, "../")
}

func safeEnvironment() map[string]string {
	allowed := map[string]bool{"HOME": true, "LANG": true, "LC_ALL": true, "PATH": true, "PWD": true, "SHELL": true, "TERM": true, "USER": true}
	out := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if ok && allowed[key] {
			out[key] = value
		}
	}
	return out
}

func environmentList(environment map[string]string) []string {
	keys := make([]string, 0, len(environment))
	for key := range environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, key+"="+environment[key])
	}
	return out
}

func snapshotToolOutput(dir string) (map[string]struct{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		seen[entry.Name()] = struct{}{}
	}
	return seen, nil
}

func generatedToolOutput(dir string, before map[string]struct{}) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	generated := make([]string, 0)
	for _, entry := range entries {
		if _, existed := before[entry.Name()]; existed || !entry.Type().IsRegular() {
			continue
		}
		generated = append(generated, filepath.ToSlash(filepath.Join("tool-output", entry.Name())))
	}
	sort.Strings(generated)
	return generated, nil
}
