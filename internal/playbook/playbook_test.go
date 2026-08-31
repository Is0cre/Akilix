package playbook

import (
	"reflect"
	"strings"
	"testing"

	containerpkg "github.com/pensuse/pensuse/internal/container"
	"github.com/pensuse/pensuse/internal/scope"
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
