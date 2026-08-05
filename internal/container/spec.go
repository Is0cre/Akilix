package container

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type Mount struct {
	Source           string
	Destination      string
	ReadOnly         bool
	OriginalEvidence bool
}
type Spec struct {
	Identity     Identity
	Arguments    []string
	Mounts       []Mount
	Network      string
	WritableRoot bool
	Environment  map[string]string
	Workdir      string
}

func Execute(ctx context.Context, runner Runner, spec Spec) (string, error) {
	args, err := spec.Args()
	if err != nil {
		return "", err
	}
	return runner.Run(ctx, "podman", args...)
}

func (s Spec) Args() ([]string, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	network := s.Network
	if network == "" {
		network = "none"
	}
	args := []string{"run", "--rm", "--pull=never", "--network=" + network, "--userns=keep-id", "--pid=private", "--ipc=private", "--uts=private", "--security-opt=no-new-privileges", "--cap-drop=ALL"}
	if !s.WritableRoot {
		args = append(args, "--read-only")
	}
	keys := make([]string, 0, len(s.Environment))
	for k := range s.Environment {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		args = append(args, "--env", k+"="+s.Environment[k])
	}
	if s.Workdir != "" {
		args = append(args, "--workdir", s.Workdir)
	}
	args = append(args, s.Identity.Image+"@"+s.Identity.Digest)
	for _, m := range s.Mounts {
		mode := "rw"
		if m.ReadOnly {
			mode = "ro"
		}
		args = append(args, "--volume", m.Source+":"+m.Destination+":"+mode)
	}
	return append(args, s.Arguments...), nil
}

func (s Spec) Validate() error {
	if !strings.HasPrefix(s.Identity.Digest, "sha256:") {
		return fmt.Errorf("container spec requires immutable image digest")
	}
	if len(s.Arguments) == 0 || strings.TrimSpace(s.Arguments[0]) == "" {
		return fmt.Errorf("container command is required")
	}
	if s.Network == "host" {
		return fmt.Errorf("host networking requires an explicit privileged policy")
	}
	for key, value := range s.Environment {
		if key == "" || strings.ContainsAny(key, "=\x00") || strings.ContainsRune(value, '\x00') {
			return fmt.Errorf("invalid container environment key %q", key)
		}
		if strings.IndexFunc(key, unicode.IsSpace) >= 0 {
			return fmt.Errorf("invalid container environment key %q", key)
		}
	}
	if s.Workdir != "" && !strings.HasPrefix(s.Workdir, "/") {
		return fmt.Errorf("container workdir must be absolute")
	}
	if s.Network != "" && s.Network != "none" && s.Network != "bridge" {
		return fmt.Errorf("unsupported container network mode %q", s.Network)
	}
	for _, m := range s.Mounts {
		if strings.TrimSpace(m.Source) == "" || strings.TrimSpace(m.Destination) == "" {
			return fmt.Errorf("container mount paths are required")
		}
		if !filepath.IsAbs(m.Source) || !filepath.IsAbs(m.Destination) || filepath.Clean(m.Source) != m.Source || filepath.Clean(m.Destination) != m.Destination || strings.ContainsAny(m.Source+m.Destination, ":\x00") || strings.Contains(m.Source, ".."+string(filepath.Separator)) || strings.Contains(m.Destination, ".."+string(filepath.Separator)) {
			return fmt.Errorf("container mount paths must be absolute, clean, and delimiter-safe")
		}
		if m.OriginalEvidence && !m.ReadOnly {
			return fmt.Errorf("original evidence mounts must be read-only")
		}
	}
	return nil
}
