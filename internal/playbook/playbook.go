package playbook

import (
	"fmt"
	"net"
	"sort"
	"strings"

	containerpkg "github.com/Is0cre/Akilix/internal/container"
	"github.com/Is0cre/Akilix/internal/scope"
)

const (
	LocalNetworkDiscovery = "local-network-discovery"
	LocalPortDiscovery    = "local-port-discovery"
)

type Attention string

const (
	AttentionInfo  Attention = "INFO"
	AttentionSafe  Attention = "SAFE"
	AttentionWarn  Attention = "WARN"
	AttentionBlock Attention = "BLOCK"
)

type Plan struct {
	Playbook       string                `json:"playbook"`
	Target         string                `json:"target"`
	Scope          scope.Decision        `json:"scope"`
	Attention      Attention             `json:"attention"`
	Image          containerpkg.Identity `json:"image"`
	Network        string                `json:"network"`
	Arguments      []string              `json:"arguments"`
	MountOriginals bool                  `json:"mount_originals"`
	MountOutput    bool                  `json:"mount_output"`
	OutputPath     string                `json:"output_path"`
	Exclusions     []string              `json:"exclusions,omitempty"`
}

type NativePlan struct {
	Playbook   string         `json:"playbook"`
	Target     string         `json:"target"`
	Scope      scope.Decision `json:"scope"`
	Attention  Attention      `json:"attention"`
	Arguments  []string       `json:"arguments"`
	Output     string         `json:"output"`
	Exclusions []string       `json:"exclusions,omitempty"`
}

// PlanNativeLocalNetworkDiscovery produces an argv-only Nmap plan whose XML
// is written to standard output and therefore captured by the invocation
// engine. It performs no DNS lookup or network operation itself.
func PlanNativeLocalNetworkDiscovery(config scope.Config, target string) (NativePlan, error) {
	_, network, err := net.ParseCIDR(strings.TrimSpace(target))
	if err != nil {
		return NativePlan{}, fmt.Errorf("local discovery target must be a CIDR: %w", err)
	}
	target = network.String()
	decision := scope.EvaluateDecision(config, target)
	if decision.Result != scope.Allow {
		return NativePlan{}, fmt.Errorf("local discovery requires explicitly allowed CIDR %s (scope: %s)", target, decision.Result)
	}
	exclusions := containedIPExclusions(config.Excludes, network)
	args := []string{"nmap", "-sn", "-n", "--reason", "-oX", "-"}
	if len(exclusions) > 0 {
		args = append(args, "--exclude", strings.Join(exclusions, ","))
	}
	args = append(args, target)
	return NativePlan{Playbook: LocalNetworkDiscovery, Target: target, Scope: decision, Attention: AttentionSafe, Arguments: args, Output: "invocation stdout (Nmap XML)", Exclusions: exclusions}, nil
}

// PlanLocalNetworkDiscovery is pure planning logic. It performs no DNS,
// registry, container, or network operation.
func PlanLocalNetworkDiscovery(config scope.Config, target string, image containerpkg.Identity) (Plan, error) {
	_, network, err := net.ParseCIDR(strings.TrimSpace(target))
	if err != nil {
		return Plan{}, fmt.Errorf("local discovery target must be a CIDR: %w", err)
	}
	target = network.String()
	decision := scope.EvaluateDecision(config, target)
	if decision.Result != scope.Allow {
		return Plan{}, fmt.Errorf("local discovery requires explicitly allowed CIDR %s (scope: %s)", target, decision.Result)
	}
	if image.Image == "" || !strings.HasPrefix(image.Digest, "sha256:") || len(image.Digest) != 71 {
		return Plan{}, fmt.Errorf("local discovery requires immutable container identity")
	}
	exclusions := containedIPExclusions(config.Excludes, network)
	args := []string{"nmap", "-sn", "-n", "--reason", "-oX", "/workbook/output/discovery.xml"}
	if len(exclusions) > 0 {
		args = append(args, "--exclude", strings.Join(exclusions, ","))
	}
	args = append(args, target)
	return Plan{Playbook: LocalNetworkDiscovery, Target: target, Scope: decision, Attention: AttentionSafe, Image: image, Network: "bridge", Arguments: args, MountOutput: true, OutputPath: "/workbook/output/discovery.xml", Exclusions: exclusions}, nil
}

// PlanLocalPortDiscovery builds a conservative, rootless Naabu CONNECT scan.
// It is pure planning logic and performs no network or container operation.
func PlanLocalPortDiscovery(config scope.Config, target string, image containerpkg.Identity) (Plan, error) {
	_, network, err := net.ParseCIDR(strings.TrimSpace(target))
	if err != nil {
		return Plan{}, fmt.Errorf("local port discovery target must be a CIDR: %w", err)
	}
	target = network.String()
	decision := scope.EvaluateDecision(config, target)
	if decision.Result != scope.Allow {
		return Plan{}, fmt.Errorf("local port discovery requires explicitly allowed CIDR %s (scope: %s)", target, decision.Result)
	}
	if image.Image == "" || !strings.HasPrefix(image.Digest, "sha256:") || len(image.Digest) != 71 {
		return Plan{}, fmt.Errorf("local port discovery requires immutable container identity")
	}
	exclusions := containedIPExclusions(config.Excludes, network)
	args := []string{
		"-host", target,
		"-scan-type", "c",
		"-top-ports", "100",
		"-rate", "100",
		"-retries", "1",
		"-timeout", "1000",
		"-disable-update-check",
		"-no-color",
		"-json",
		"-output", "/workbook/output/ports.jsonl",
	}
	if len(exclusions) > 0 {
		args = append(args, "-exclude-hosts", strings.Join(exclusions, ","))
	}
	return Plan{Playbook: LocalPortDiscovery, Target: target, Scope: decision, Attention: AttentionSafe, Image: image, Network: "bridge", Arguments: args, MountOutput: true, OutputPath: "/workbook/output/ports.jsonl", Exclusions: exclusions}, nil
}

func containedIPExclusions(exclusions []string, target *net.IPNet) []string {
	var out []string
	for _, value := range exclusions {
		if ip := net.ParseIP(value); ip != nil && target.Contains(ip) {
			out = append(out, ip.String())
			continue
		}
		if _, subnet, err := net.ParseCIDR(value); err == nil && target.Contains(subnet.IP) && containsNetwork(target, subnet) {
			out = append(out, subnet.String())
		}
	}
	sort.Strings(out)
	return out
}

func containsNetwork(parent, child *net.IPNet) bool {
	ones, bits := child.Mask.Size()
	last := append(net.IP(nil), child.IP...)
	for bit := ones; bit < bits; bit++ {
		last[bit/8] |= 1 << uint(7-bit%8)
	}
	return parent.Contains(last)
}
