package profile

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

const VerificationSchema = "akilix.profile-verification.v1"

type CommandRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}
type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

type ComponentState struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Present bool   `json:"present"`
	Detail  string `json:"detail,omitempty"`
}
type Verification struct {
	Schema     string           `json:"schema"`
	ID         string           `json:"verification_id"`
	ProfileID  string           `json:"profile_id"`
	VerifiedAt time.Time        `json:"verified_at"`
	Ready      bool             `json:"ready"`
	Components []ComponentState `json:"components"`
}

// Verify performs read-only local package/image queries. It never refreshes
// repositories, pulls images, installs packages, or changes services.
func Verify(ctx context.Context, manifest Manifest, runner CommandRunner, now time.Time) (Verification, error) {
	if err := manifest.Validate(); err != nil {
		return Verification{}, err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id, err := verificationID(now)
	if err != nil {
		return Verification{}, err
	}
	result := Verification{Schema: VerificationSchema, ID: id, ProfileID: manifest.ID, VerifiedAt: now.UTC(), Ready: true, Components: []ComponentState{}}
	for _, name := range manifest.RPM {
		out, queryErr := runner.Run(ctx, "rpm", "-q", "--quiet", "--", name)
		state := ComponentState{Kind: "rpm", Name: name, Present: queryErr == nil}
		if queryErr != nil {
			state.Detail = boundedDetail(out, queryErr)
			result.Ready = false
		}
		result.Components = append(result.Components, state)
	}
	for _, name := range manifest.Containers {
		image := name
		if !strings.Contains(image, "/") {
			image = "localhost/" + image
		}
		out, queryErr := runner.Run(ctx, "podman", "image", "inspect", "--format", "{{.Digest}}", image)
		digest := strings.TrimSpace(string(out))
		present := queryErr == nil && strings.HasPrefix(digest, "sha256:") && len(digest) == 71
		state := ComponentState{Kind: "oci", Name: image, Present: present}
		if present {
			state.Detail = digest
		} else {
			state.Detail = boundedDetail(out, queryErr)
			result.Ready = false
		}
		result.Components = append(result.Components, state)
	}
	return result, nil
}

func RecordVerification(stateDir string, verification Verification) (string, error) {
	if verification.Schema != VerificationSchema || verification.ID == "" || verification.ProfileID == "" || verification.VerifiedAt.IsZero() {
		return "", fmt.Errorf("invalid profile verification")
	}
	dir := filepath.Join(stateDir, "platform", "profile-verifications")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("invalid profile verification directory")
	}
	data, err := json.MarshalIndent(verification, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	path := filepath.Join(dir, verification.ID+".json")
	tmp, err := os.CreateTemp(dir, ".verify-")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", err
	}
	if err := os.Link(tmpPath, path); err != nil {
		return "", err
	}
	if err := os.Remove(tmpPath); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	d, err := os.Open(dir)
	if err != nil {
		return "", err
	}
	err = d.Sync()
	_ = d.Close()
	return path, err
}

func boundedDetail(out []byte, err error) string {
	value := strings.TrimSpace(string(out))
	if value == "" && err != nil {
		value = err.Error()
	}
	if len(value) > 512 {
		value = value[:512]
	}
	return value
}
func verificationID(now time.Time) (string, error) {
	var b [10]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("PV-%d-%s", now.UnixMilli(), hex.EncodeToString(b[:])), nil
}
