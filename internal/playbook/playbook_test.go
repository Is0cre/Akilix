package playbook

import (
	"reflect"
	"strings"
	"testing"

	containerpkg "github.com/Is0cre/Akilix/internal/container"
	"github.com/Is0cre/Akilix/internal/scope"
)

func TestPlanLocalDiscoveryRequiresExplicitCIDRScope(t *testing.T) {
	identity := containerpkg.Identity{Image: "example/discovery:1", Digest: "sha256:" + strings.Repeat("a", 64)}
	config := scope.Config{Includes: []string{"192.168.50.0/24"}, Excludes: []string{"192.168.50.1", "10.0.0.0/8"}}
	plan, err := PlanLocalNetworkDiscovery(config, "192.168.50.23/24", identity)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Target != "192.168.50.0/24" || plan.Network != "bridge" || !plan.MountOutput || plan.MountOriginals || plan.Attention != AttentionSafe {
		t.Fatalf("unsafe plan: %+v", plan)
	}
	want := []string{"nmap", "-sn", "-n", "--reason", "-oX", "/workbook/output/discovery.xml", "--exclude", "192.168.50.1", "192.168.50.0/24"}
	if !reflect.DeepEqual(plan.Arguments, want) {
		t.Fatalf("args=%v want=%v", plan.Arguments, want)
	}
}

func TestPlanLocalDiscoveryBlocksUnknownAndNonCIDR(t *testing.T) {
	identity := containerpkg.Identity{Image: "example/discovery:1", Digest: "sha256:" + strings.Repeat("a", 64)}
	config := scope.Config{Includes: []string{"192.168.50.0/24"}}
	for _, target := range []string{"10.0.0.0/24", "192.168.50.1"} {
		if _, err := PlanLocalNetworkDiscovery(config, target, identity); err == nil {
			t.Fatalf("accepted %s", target)
		}
	}
}

func TestPlanNativeLocalNetworkDiscoveryUsesCapturedXML(t *testing.T) {
	config := scope.Config{Includes: []string{"192.168.50.0/24"}, Excludes: []string{"192.168.50.1"}}
	plan, err := PlanNativeLocalNetworkDiscovery(config, "192.168.50.13/24")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"nmap", "-sn", "-n", "--reason", "-oX", "-", "--exclude", "192.168.50.1", "192.168.50.0/24"}
	if !reflect.DeepEqual(plan.Arguments, want) || plan.Scope.Result != scope.Allow {
		t.Fatalf("plan=%+v", plan)
	}
}

func TestPlanNativeLocalNetworkDiscoveryRejectsUndeclaredTarget(t *testing.T) {
	if _, err := PlanNativeLocalNetworkDiscovery(scope.Config{}, "192.168.50.0/24"); err == nil {
		t.Fatal("undeclared target accepted")
	}
}

func TestPlanLocalPortDiscoveryUsesConservativeRootlessFlags(t *testing.T) {
	identity := containerpkg.Identity{Image: "example/naabu:1", Digest: "sha256:" + strings.Repeat("b", 64)}
	config := scope.Config{Includes: []string{"192.168.50.0/24"}, Excludes: []string{"192.168.50.9", "10.0.0.0/8"}}
	plan, err := PlanLocalPortDiscovery(config, "192.168.50.23/24", identity)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-host", "192.168.50.0/24", "-scan-type", "c", "-top-ports", "100", "-rate", "100", "-retries", "1", "-timeout", "1000", "-disable-update-check", "-no-color", "-json", "-output", "/workbook/output/ports.jsonl", "-exclude-hosts", "192.168.50.9"}
	if !reflect.DeepEqual(plan.Arguments, want) {
		t.Fatalf("args=%v want=%v", plan.Arguments, want)
	}
	if plan.Playbook != LocalPortDiscovery || plan.OutputPath != "/workbook/output/ports.jsonl" || plan.Network != "bridge" || !plan.MountOutput || plan.MountOriginals {
		t.Fatalf("unsafe plan: %+v", plan)
	}
}

func TestPlanLocalPortDiscoveryBlocksUnknownAndMutableImage(t *testing.T) {
	good := containerpkg.Identity{Image: "example/naabu:1", Digest: "sha256:" + strings.Repeat("b", 64)}
	config := scope.Config{Includes: []string{"192.168.50.0/24"}}
	if _, err := PlanLocalPortDiscovery(config, "10.0.0.0/24", good); err == nil {
		t.Fatal("accepted unknown target")
	}
	if _, err := PlanLocalPortDiscovery(config, "192.168.50.0/24", containerpkg.Identity{Image: "example/naabu:latest"}); err == nil {
		t.Fatal("accepted mutable image identity")
	}
}
