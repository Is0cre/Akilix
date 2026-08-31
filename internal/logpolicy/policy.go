package logpolicy

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const Schema = "akilix.logging.v1"

type Policy struct {
	Schema                string `json:"schema"`
	CommandMetadata       bool   `json:"command_metadata"`
	CommandArguments      bool   `json:"command_arguments"`
	ContainerMetadata     bool   `json:"container_metadata"`
	EvidenceHashing       bool   `json:"evidence_hashing"`
	StdoutCapture         bool   `json:"stdout_capture"`
	StderrCapture         bool   `json:"stderr_capture"`
	GeneratedFileTracking bool   `json:"generated_file_tracking"`
	PacketMetadata        bool   `json:"packet_metadata"`
	TerminalRecording     bool   `json:"terminal_recording"`
}

func Default() Policy {
	return Policy{
		Schema: Schema, CommandMetadata: true, CommandArguments: true,
		ContainerMetadata: true, EvidenceHashing: true, StdoutCapture: true,
		StderrCapture: true, GeneratedFileTracking: true,
	}
}

func (p Policy) Validate() error {
	if p.Schema != Schema {
		return fmt.Errorf("unsupported logging policy schema %q", p.Schema)
	}
	return nil
}

func Render(p Policy) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	return fmt.Sprintf("schema: %s\ncommand_metadata: %t\ncommand_arguments: %t\ncontainer_metadata: %t\nevidence_hashing: %t\nstdout_capture: %t\nstderr_capture: %t\ngenerated_file_tracking: %t\npacket_metadata: %t\nterminal_recording: %t\n",
		p.Schema, p.CommandMetadata, p.CommandArguments, p.ContainerMetadata,
		p.EvidenceHashing, p.StdoutCapture, p.StderrCapture,
		p.GeneratedFileTracking, p.PacketMetadata, p.TerminalRecording), nil
}

func Load(workbookRoot string) (Policy, error) {
	b, err := os.ReadFile(filepath.Join(workbookRoot, "logging.yaml"))
	if err != nil {
		return Policy{}, err
	}
	values := map[string]string{}
	for number, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			return Policy{}, fmt.Errorf("invalid logging policy line %d", number+1)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if _, exists := values[key]; exists {
			return Policy{}, fmt.Errorf("duplicate logging policy key %q", key)
		}
		values[key] = value
	}
	known := map[string]bool{"schema": true, "command_metadata": true, "command_arguments": true, "container_metadata": true, "evidence_hashing": true, "stdout_capture": true, "stderr_capture": true, "generated_file_tracking": true, "packet_metadata": true, "terminal_recording": true}
	for key := range values {
		if !known[key] {
			return Policy{}, fmt.Errorf("unknown logging policy key %q", key)
		}
	}
	boolean := func(key string) (bool, error) {
		value, ok := values[key]
		if !ok {
			return false, fmt.Errorf("missing logging policy key %q", key)
		}
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return false, fmt.Errorf("invalid logging policy value for %q", key)
		}
		return parsed, nil
	}
	p := Policy{Schema: values["schema"]}
	fields := []struct {
		key string
		dst *bool
	}{{"command_metadata", &p.CommandMetadata}, {"command_arguments", &p.CommandArguments}, {"container_metadata", &p.ContainerMetadata}, {"evidence_hashing", &p.EvidenceHashing}, {"stdout_capture", &p.StdoutCapture}, {"stderr_capture", &p.StderrCapture}, {"generated_file_tracking", &p.GeneratedFileTracking}, {"packet_metadata", &p.PacketMetadata}, {"terminal_recording", &p.TerminalRecording}}
	for _, field := range fields {
		value, err := boolean(field.key)
		if err != nil {
			return Policy{}, err
		}
		*field.dst = value
	}
	return p, p.Validate()
}
