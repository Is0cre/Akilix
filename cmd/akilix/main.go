package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Is0cre/Akilix/internal/acquire"
	"github.com/Is0cre/Akilix/internal/activity"
	"github.com/Is0cre/Akilix/internal/completion"
	"github.com/Is0cre/Akilix/internal/config"
	containerpkg "github.com/Is0cre/Akilix/internal/container"
	"github.com/Is0cre/Akilix/internal/evidence"
	greeterpkg "github.com/Is0cre/Akilix/internal/greeter"
	"github.com/Is0cre/Akilix/internal/invocation"
	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/logpolicy"
	playbookpkg "github.com/Is0cre/Akilix/internal/playbook"
	profilepkg "github.com/Is0cre/Akilix/internal/profile"
	repositorypkg "github.com/Is0cre/Akilix/internal/repository"
	"github.com/Is0cre/Akilix/internal/scope"
	"github.com/Is0cre/Akilix/internal/statusbar"
	tuipkg "github.com/Is0cre/Akilix/internal/tui"
	"github.com/Is0cre/Akilix/internal/version"
	"github.com/Is0cre/Akilix/internal/workbook"
	"github.com/Is0cre/Akilix/internal/workbookview"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return runTUIHome(os.Stdin, stdout, stderr)
	}
	if args[0] == "workbook" {
		return runWorkbook(args[1:], stdout, stderr)
	}
	if args[0] == "evidence" {
		return runEvidence(args[1:], stdout, stderr)
	}
	if args[0] == "acquire" {
		return runAcquire(args[1:], stdout, stderr)
	}
	if args[0] == "device" {
		return runDevice(args[1:], stdout, stderr)
	}
	if args[0] == "scope" {
		return runScope(args[1:], stdout, stderr)
	}
	if args[0] == "logging" {
		return runLogging(args[1:], stdout, stderr)
	}
	if args[0] == "run" {
		return runCommand(args[1:], stdout, stderr)
	}
	if args[0] == "container" {
		return runContainer(args[1:], stdout, stderr)
	}
	if args[0] == "profile" {
		return runProfile(args[1:], stdout, stderr)
	}
	if args[0] == "repository" {
		return runRepository(args[1:], stdout, stderr)
	}
	if args[0] == "config" {
		return runConfig(args[1:], stdout, stderr)
	}
	if args[0] == "completion" {
		return runCompletion(args[1:], stdout, stderr)
	}
	if args[0] == "bar" {
		return runBar(args[1:], stdout, stderr)
	}
	if args[0] == "greeter" {
		return runGreeter(args[1:], stdout, stderr)
	}
	if args[0] == "tui" {
		if len(args) == 1 {
			return runTUIHome(os.Stdin, stdout, stderr)
		}
		if len(args) > 3 || len(args) == 3 && args[2] != "--no-color" {
			fmt.Fprintln(stderr, "usage: akilix tui [WORKBOOK] [--no-color]")
			return 2
		}
		return runTUISession(bufio.NewScanner(os.Stdin), effectiveWorkbookRoot(), args[1], len(args) != 3, stdout, stderr)
	}
	if args[0] != "version" {
		fmt.Fprintln(stderr, "usage: akilix version [--json] | workbook ... | scope ... | evidence ... | acquire ... | device ... | run ... | container ... | bar ... | greeter ...")
		return 2
	}
	info := version.Current()
	if len(args) == 2 && args[1] == "--json" {
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: akilix version [--json]")
		return 2
	}
	fmt.Fprintf(stdout, "%s %s\nBase: %s\nArchitecture: %s\n", info.Name, info.Version, info.Base, info.Architecture)
	return 0
}

func runGreeter(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 || args[0] != "preflight" || len(args) > 2 || len(args) == 2 && args[1] != "--no-color" && args[1] != "--json" {
		fmt.Fprintln(stderr, "usage: akilix greeter preflight [--no-color|--json]")
		return 2
	}
	snapshot := greeterpkg.Inspect("/")
	if len(args) == 2 && args[1] == "--json" {
		data, err := json.MarshalIndent(snapshot, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	fmt.Fprintln(stdout, greeterpkg.Render(snapshot, len(args) != 2))
	return 0
}

func runDevice(args []string, stdout, stderr io.Writer) int {
	usage := "usage: akilix device trust add DEVICE [LABEL] | device trust list [--json] | device trust remove TRUST_ID"
	if len(args) < 2 || args[0] != "trust" {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	path := filepath.Join(effectiveStateDir(), "devices", "trusted.json")
	registry, err := acquire.LoadTrust(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	switch args[1] {
	case "list":
		if len(args) > 3 || len(args) == 3 && args[2] != "--json" {
			fmt.Fprintln(stderr, usage)
			return 2
		}
		if len(args) == 3 {
			data, err := json.MarshalIndent(registry, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		if len(registry.Entries) == 0 {
			fmt.Fprintln(stdout, "No trusted devices")
			return 0
		}
		for _, entry := range registry.Entries {
			fmt.Fprintf(stdout, "%s  %-12s %-20s %s %s\n", entry.ID, entry.Kind, entry.Label, entry.Vendor, entry.Model)
		}
		return 0
	case "add":
		if len(args) != 3 && len(args) != 4 {
			fmt.Fprintln(stderr, usage)
			return 2
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		report, err := acquire.Inspect(ctx, acquire.ExecRunner{}, time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		device, err := acquire.FindWholeDisk(report, args[2])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		device, err = acquire.EnrichUSB(ctx, acquire.ExecRunner{}, device, usbIDsPath())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		label := ""
		if len(args) == 4 {
			label = args[3]
		}
		entry, err := registry.Add(device, label, time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if err := acquire.SaveTrust(path, registry); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "trusted %s as %s\n", device.Path, entry.ID)
		fmt.Fprintln(stdout, "Trust does not make storage writable or disable acquisition protection.")
		return 0
	case "remove":
		if len(args) != 3 {
			fmt.Fprintln(stderr, usage)
			return 2
		}
		entry, err := registry.Remove(args[2], time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if err := acquire.SaveTrust(path, registry); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "removed trust for %s (%s)\n", entry.ID, entry.Label)
		return 0
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
}

func runAcquire(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 || args[0] != "inspect" && args[0] != "record" && args[0] != "protect" && args[0] != "identify" && args[0] != "image" && args[0] != "verify" && args[0] != "status" {
		fmt.Fprintln(stderr, "usage: akilix acquire inspect [--json] | acquire record WORKBOOK [--json] | acquire identify WORKBOOK DEVICE [--json] | acquire protect WORKBOOK DEVICE [--json] | acquire image WORKBOOK DEVICE OUTPUT [--json] | acquire verify WORKBOOK OPERATION_ID [--json] | acquire status WORKBOOK [OPERATION_ID] [--json]")
		return 2
	}
	jsonOutput := false
	if args[0] == "inspect" {
		if len(args) > 2 || len(args) == 2 && args[1] != "--json" {
			fmt.Fprintln(stderr, "usage: akilix acquire inspect [--json]")
			return 2
		}
		jsonOutput = len(args) == 2
	} else if args[0] == "record" {
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix acquire record WORKBOOK [--json]")
			return 2
		}
		jsonOutput = len(args) == 3
	} else if args[0] == "image" {
		if len(args) != 4 && !(len(args) == 5 && args[4] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix acquire image WORKBOOK DEVICE OUTPUT [--json]")
			return 2
		}
		jsonOutput = len(args) == 5
	} else if args[0] == "status" {
		if len(args) < 2 || len(args) > 4 || len(args) == 4 && args[3] != "--json" {
			fmt.Fprintln(stderr, "usage: akilix acquire status WORKBOOK [OPERATION_ID] [--json]")
			return 2
		}
		jsonOutput = len(args) == 3 && args[2] == "--json" || len(args) == 4
	} else {
		if len(args) != 3 && !(len(args) == 4 && args[3] == "--json") {
			fmt.Fprintf(stderr, "usage: akilix acquire %s WORKBOOK DEVICE [--json]\n", args[0])
			return 2
		}
		jsonOutput = len(args) == 4
	}
	var metadata workbook.Metadata
	var workbookRoot string
	if args[0] == "record" || args[0] == "protect" || args[0] == "identify" || args[0] == "image" || args[0] == "verify" || args[0] == "status" {
		root := effectiveWorkbookRoot()
		var err error
		metadata, err = workbook.Open(root, args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if metadata.Status != "open" {
			fmt.Fprintln(stderr, "workbook is closed")
			return 1
		}
		workbookRoot = filepath.Join(root, args[1])
	}
	if args[0] == "status" {
		opID := ""
		if len(args) >= 3 && args[2] != "--json" {
			opID = args[2]
		}
		states, err := acquire.ImageStatus(workbookRoot, opID)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if jsonOutput {
			data, marshalErr := json.MarshalIndent(states, "", "  ")
			if marshalErr != nil {
				fmt.Fprintln(stderr, marshalErr)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, state := range states {
			fmt.Fprintf(stdout, "%s\t%-9s\trecovery=%t\t%s\n", state.OperationID, state.State, state.RecoveryRequired, state.Destination)
		}
		return 0
	}
	if args[0] == "verify" {
		verifyCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		record, match, err := acquire.VerifyImage(verifyCtx, workbookRoot, metadata.ID, args[2], time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if jsonOutput {
			data, marshalErr := json.MarshalIndent(record, "", "  ")
			if marshalErr != nil {
				fmt.Fprintln(stderr, marshalErr)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
		} else {
			fmt.Fprintf(stdout, "acquisition %s verification: %s\nSHA-256 %s\n", record.OperationID, record.Verification, record.SHA256)
		}
		if !match {
			return 1
		}
		return 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	now := time.Now().UTC()
	report, err := acquire.Inspect(ctx, acquire.ExecRunner{}, now)
	cancel()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	registry, err := acquire.LoadTrust(filepath.Join(effectiveStateDir(), "devices", "trusted.json"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	acquire.ApplyTrust(&report, registry)
	if args[0] == "image" {
		device, err := acquire.Candidate(report, args[2])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if !device.ReadOnly {
			fmt.Fprintln(stderr, "refusing imaging until kernel read-only state is verified; run akilix acquire protect first")
			return 1
		}
		verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 10*time.Second)
		verified, err := acquire.SetReadOnly(verifyCtx, acquire.ExecRunner{}, device)
		verifyCancel()
		if err != nil || !verified.KernelReadOnly {
			fmt.Fprintln(stderr, "kernel read-only revalidation failed immediately before imaging:", err)
			return 1
		}
		imageCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		record, path, err := acquire.Image(imageCtx, workbookRoot, metadata.ID, device, args[3], now)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if jsonOutput {
			data, marshalErr := json.MarshalIndent(record, "", "  ")
			if marshalErr != nil {
				fmt.Fprintln(stderr, marshalErr)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
		} else {
			fmt.Fprintf(stdout, "acquired %s to %s\nSHA-256 %s\n", device.Path, path, record.SHA256)
			fmt.Fprintln(stdout, "Software read-only protection is not a hardware forensic write blocker.")
		}
		return 0
	}
	if args[0] == "record" {
		record, path, err := acquire.RecordInspection(workbookRoot, metadata.ID, report, now)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if jsonOutput {
			data, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		fmt.Fprintf(stdout, "recorded passive hardware inventory %s\n%s\n", record.ID, path)
		return 0
	}
	if args[0] == "identify" {
		device, err := acquire.Candidate(report, args[2])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		inventory, _, err := acquire.RecordInspection(workbookRoot, metadata.ID, report, now)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		identification, err := acquire.Identify(ctx, acquire.ExecResultRunner{}, device, inventory.ID, time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		record, path, err := acquire.RecordIdentification(workbookRoot, metadata.ID, identification, time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if jsonOutput {
			data, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
		} else {
			fmt.Fprintf(stdout, "recorded passive identity %s for %s\n%s\n", record.ID, device.Path, path)
			for _, command := range record.Commands {
				fmt.Fprintf(stdout, "%-10s %s (exit %d)\n", command.Tool, command.Status, command.ExitCode)
			}
		}
		for _, command := range record.Commands {
			if command.Status == "UNAVAILABLE" || command.Status == "TOOL_ERROR" || command.Status == "INVALID_OUTPUT" {
				return 1
			}
		}
		return 0
	}
	if args[0] == "protect" {
		device, err := acquire.Candidate(report, args[2])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		operationID, err := acquire.NewOperationID(now)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if _, _, err := acquire.RecordProtectionEvent(workbookRoot, metadata.ID, operationID, "REQUESTED", device, acquire.ProtectionResult{}, nil, now); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		result, protectErr := acquire.SetReadOnly(ctx, acquire.ExecRunner{}, device)
		phase := "APPLIED"
		if protectErr != nil {
			phase = "FAILED"
		}
		record, path, recordErr := acquire.RecordProtectionEvent(workbookRoot, metadata.ID, operationID, phase, device, result, protectErr, time.Now().UTC())
		if protectErr != nil {
			if recordErr != nil {
				fmt.Fprintf(stderr, "%v; additionally failed to record failure: %v\n", protectErr, recordErr)
			} else {
				fmt.Fprintln(stderr, protectErr)
			}
			return 1
		}
		if recordErr != nil {
			fmt.Fprintf(stderr, "kernel read-only state verified, but completion provenance failed: %v\n", recordErr)
			return 1
		}
		if jsonOutput {
			data, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		fmt.Fprintf(stdout, "kernel read-only state verified for %s\noperation %s\n%s\n", device.Path, operationID, path)
		fmt.Fprintln(stdout, "Software read-only protection is not a hardware forensic write blocker.")
		return 0
	}
	if jsonOutput {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	fmt.Fprintln(stdout, "Passive hardware inspection — no device state changed")
	if len(report.Devices) == 0 {
		fmt.Fprintln(stdout, "No whole-disk block devices found")
		return 0
	}
	for _, device := range report.Devices {
		role := "CANDIDATE"
		if device.SystemDisk {
			role = "SYSTEM"
		}
		trust := "UNKNOWN"
		if device.Trusted {
			trust = "TRUSTED:" + device.TrustID
		}
		fmt.Fprintf(stdout, "%s  %-14s %8s  RO=%-3t RM=%-3t %-5s %s %s  %s\n", role, device.Path, humanBytes(device.SizeBytes), device.ReadOnly, device.Removable, device.Transport, device.Vendor, device.Model, trust)
		fmt.Fprintf(stdout, "        serial=%s wwn=%s mounted=%t", emptyDash(device.Serial), emptyDash(device.WWN), device.Mounted)
		if len(device.Mountpoints) != 0 {
			fmt.Fprintf(stdout, " at %s", strings.Join(device.Mountpoints, ","))
		}
		fmt.Fprintln(stdout)
		for _, partition := range device.Partitions {
			fmt.Fprintf(stdout, "        PART %-14s %8s  %-8s uuid=%s mounted=%t", partition.Path, humanBytes(partition.SizeBytes), emptyDash(partition.Filesystem), emptyDash(partition.UUID), partition.Mounted)
			if len(partition.Mountpoints) != 0 {
				fmt.Fprintf(stdout, " at %s", strings.Join(partition.Mountpoints, ","))
			}
			fmt.Fprintln(stdout)
		}
	}
	return 0
}

func humanBytes(size uint64) string {
	const unit = uint64(1024)
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	value := float64(size)
	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	for _, suffix := range units {
		value /= 1024
		if value < 1024 || suffix == units[len(units)-1] {
			return fmt.Sprintf("%.1f%s", value, suffix)
		}
	}
	return fmt.Sprintf("%dB", size)
}

func emptyDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func runTUIHome(stdin io.Reader, stdout, stderr io.Writer) int {
	scanner := bufio.NewScanner(stdin)
	root := effectiveWorkbookRoot()
	items, err := workbook.List(root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(items) == 0 {
		if !isTerminalWriter(stdout) || !isTerminalFile(os.Stdin) {
			fmt.Fprintln(stdout, "Akilix has no workbooks yet.\n\nCreate one with:\n  akilix workbook create NAME")
			return 0
		}
		name, ok := promptCreateFirstWorkbook(scanner, stdout)
		if !ok {
			return 0
		}
		created, err := workbook.Create(root, name, timeNow())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "created %s (%s)\n", created.Name, created.ID)
		items = []workbook.Metadata{created}
	}
	selected := items[0].Name
	if len(items) > 1 {
		fmt.Fprintln(stdout, "Akilix workbooks")
		for i, item := range items {
			fmt.Fprintf(stdout, "  %d  %-24s %s\n", i+1, item.Name, strings.ToUpper(item.Status))
		}
		fmt.Fprint(stdout, "Select workbook: ")
		if !scanner.Scan() {
			fmt.Fprintln(stderr, "workbook selection cancelled")
			return 1
		}
		choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || choice < 1 || choice > len(items) {
			fmt.Fprintln(stderr, "workbook selection cancelled")
			return 1
		}
		selected = items[choice-1].Name
	}
	return runTUISession(scanner, root, selected, os.Getenv("NO_COLOR") == "", stdout, stderr)
}

func promptCreateFirstWorkbook(scanner *bufio.Scanner, stdout io.Writer) (string, bool) {
	fmt.Fprint(stdout, "Akilix has no workbooks yet. Create one now? [Y/n]: ")
	if !scanner.Scan() {
		return "", false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if answer != "" && answer != "y" && answer != "yes" {
		return "", false
	}
	fmt.Fprint(stdout, "Workbook name: ")
	if !scanner.Scan() {
		return "", false
	}
	name := strings.TrimSpace(scanner.Text())
	return name, name != ""
}

func runTUISession(scanner *bufio.Scanner, root, selected string, color bool, stdout, stderr io.Writer) int {
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		if err := statusbar.Activate(runtimeDir, root, selected); err != nil {
			fmt.Fprintln(stderr, "activate workbook status:", err)
		}
	}
	for {
		clearTUIScreen(stdout)
		overview, err := workbookview.Build(root, selected)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		config, err := scope.Load(filepath.Join(root, selected))
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		dashboard := tuipkg.Render(overview, config, color)
		var action string
		if isTerminalWriter(stdout) && isTerminalFile(os.Stdin) {
			action, err = selectInteractiveAction(root, selected, overview.ID, dashboard, color, stdout)
			if err != nil {
				fmt.Fprintln(stderr, "select action:", err)
				return 1
			}
		} else {
			fmt.Fprint(stdout, dashboard)
			fmt.Fprint(stdout, tuipkg.RenderActions(color))
			fmt.Fprint(stdout, "Select action > ")
			if !scanner.Scan() {
				return 0
			}
			action = strings.ToLower(strings.TrimSpace(scanner.Text()))
		}
		if action == "q" || action == "quit" {
			return 0
		}
		if action == "l" {
			if err := launchWorkbookLog(selected); err != nil {
				fmt.Fprintln(stderr, err)
			}
			continue
		}
		if action == "n" || action == "p" {
			if overview.Status != "open" {
				fmt.Fprintln(stderr, "workbook is closed")
				continue
			}
			prompt, tool := "Network", "Nmap"
			if action == "p" {
				prompt, tool = "Port", "Naabu"
			}
			fmt.Fprintf(stdout, "%s discovery CIDR: ", prompt)
			if !scanner.Scan() {
				return 0
			}
			target := strings.TrimSpace(scanner.Text())
			fmt.Fprintf(stdout, "Local %s image (already pulled): ", tool)
			if !scanner.Scan() {
				return 0
			}
			image := strings.TrimSpace(scanner.Text())
			identity, err := containerpkg.Resolve(context.Background(), containerpkg.PodmanRunner{}, image)
			if err != nil {
				fmt.Fprintln(stderr, err)
				continue
			}
			var plan playbookpkg.Plan
			if action == "n" {
				plan, err = playbookpkg.PlanLocalNetworkDiscovery(config, target, identity)
			} else {
				plan, err = playbookpkg.PlanLocalPortDiscovery(config, target, identity)
			}
			if err != nil {
				fmt.Fprintln(stderr, err)
				continue
			}
			fmt.Fprintf(stdout, "\nPLAYBOOK  %s\nSCOPE     %s via %s\nIMAGE     %s@%s\nNETWORK   %s\nOUTPUT    %s\nCOMMAND   %s\n", plan.Playbook, plan.Scope.Result, plan.Scope.Rule, plan.Image.Image, plan.Image.Digest, plan.Network, plan.OutputPath, strings.Join(plan.Arguments, " "))
			fmt.Fprint(stdout, "Type RUN to execute this scoped discovery: ")
			if !scanner.Scan() || scanner.Text() != "RUN" {
				fmt.Fprintln(stdout, "cancelled")
				continue
			}
			spec := containerpkg.Spec{Identity: plan.Image, Arguments: plan.Arguments, Network: plan.Network, InvocationOutput: true}
			record, err := invocation.RunContainer(context.Background(), filepath.Join(root, selected), overview.ID, spec, time.Now, invocation.Options{ScopeResult: string(plan.Scope.Result), ScopeTarget: plan.Target, OnStarted: func(id string) {
				if launchErr := launchInvocationWorkspace(selected, plan.Playbook, id); launchErr != nil {
					fmt.Fprintln(stderr, "open invocation workspace:", launchErr)
				}
			}})
			if err != nil {
				fmt.Fprintf(stderr, "invocation %s failed: %v\n", record.ID, err)
				continue
			}
			fmt.Fprintf(stdout, "discovery complete: invocation %s\n", record.ID)
			if action == "p" {
				log, journalErr := journal.Open(filepath.Join(root, selected))
				if journalErr != nil {
					fmt.Fprintln(stderr, "open workbook journal:", journalErr)
					continue
				}
				portsPath := filepath.Join(root, selected, "artifacts", "derived", record.ID, "ports.jsonl")
				found, dropped, journalErr := playbookpkg.IngestNaabuJournal(portsPath, record.ID, config, log, time.Now)
				if journalErr != nil {
					fmt.Fprintln(stderr, "ingest Naabu journal:", journalErr)
					continue
				}
				fmt.Fprintf(stdout, "journaled %d in-scope ports; dropped %d out-of-scope results\n", found, dropped)
			}
			continue
		}
		if action != "a" && action != "x" {
			fmt.Fprintln(stderr, "unknown action")
			continue
		}
		if overview.Status != "open" {
			fmt.Fprintln(stderr, "workbook is closed")
			continue
		}
		fmt.Fprint(stdout, "Target (IP, CIDR, hostname, domain, or URL): ")
		if !scanner.Scan() {
			return 0
		}
		workbookRoot := filepath.Join(root, selected)
		log, err := journal.Open(workbookRoot)
		if err != nil {
			fmt.Fprintln(stderr, err)
			continue
		}
		if _, err := scope.AddRecorded(workbookRoot, overview.ID, strings.TrimSpace(scanner.Text()), action == "x", log, time.Now().UTC()); err != nil {
			fmt.Fprintln(stderr, err)
			continue
		}
		fmt.Fprintln(stdout, "scope updated")
	}
}

func clearTUIScreen(stdout io.Writer) {
	if !isTerminalWriter(stdout) {
		return
	}
	fmt.Fprint(stdout, "\033[2J\033[H")
}

func isTerminalWriter(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	return ok && isTerminalFile(file)
}

func isTerminalFile(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func selectInteractiveAction(root, workbookName, workbookID, prefix string, color bool, stdout io.Writer) (string, error) {
	workbookRoot := filepath.Join(root, workbookName)
	log, err := journal.Open(workbookRoot)
	if err != nil {
		return "", err
	}
	refresh := func() (string, error) {
		overview, err := workbookview.Build(root, workbookName)
		if err != nil {
			return "", err
		}
		config, err := scope.Load(workbookRoot)
		if err != nil {
			return "", err
		}
		return tuipkg.Render(overview, config, color), nil
	}
	addScope := func(target string) error {
		_, err := scope.AddRecorded(workbookRoot, workbookID, target, false, log, time.Now().UTC())
		return err
	}
	runScan := func(ctx context.Context, action tuipkg.Action, target string) (string, error) {
		image := "localhost/local-nmap"
		if action == tuipkg.PortDiscovery {
			image = "localhost/local-naabu"
		}
		currentScope, err := scope.Load(workbookRoot)
		if err != nil {
			return image, err
		}
		if action == tuipkg.NetworkDiscovery || action == tuipkg.PortDiscovery {
			var plan playbookpkg.NativePlan
			if action == tuipkg.NetworkDiscovery {
				plan, err = playbookpkg.PlanNativeLocalNetworkDiscovery(currentScope, target)
			} else {
				plan, err = playbookpkg.PlanNativeLocalPortDiscovery(currentScope, target)
			}
			if err != nil {
				return "nmap", err
			}
			record, err := invocation.RunWithOptions(ctx, workbookRoot, workbookID, plan.Arguments, time.Now, invocation.Options{ScopeResult: string(plan.Scope.Result), ScopeTarget: plan.Target, OnStarted: func(id string) { _ = launchInvocationWorkspace(workbookName, plan.Playbook, id) }})
			if err != nil {
				if ctx.Err() != nil {
					return "nmap", ctx.Err()
				}
				return "nmap", fmt.Errorf("invocation %s failed: %w", record.ID, err)
			}
			stdoutPath := filepath.Join(workbookRoot, filepath.FromSlash(record.Stdout))
			if _, _, err := playbookpkg.IngestNmapJournal(stdoutPath, record.ID, currentScope, log, time.Now); err != nil {
				return "nmap", fmt.Errorf("ingest invocation %s output: %w", record.ID, err)
			}
			return "nmap", nil
		}
		identity, err := containerpkg.Resolve(ctx, containerpkg.PodmanRunner{}, image)
		if err != nil {
			return image, fmt.Errorf("resolve local OCI image %s: %w", image, err)
		}
		var plan playbookpkg.Plan
		plan, err = playbookpkg.PlanLocalPortDiscovery(currentScope, target, identity)
		if err != nil {
			return image, err
		}
		spec := containerpkg.Spec{Identity: plan.Image, Arguments: plan.Arguments, Network: plan.Network, InvocationOutput: true}
		record, err := invocation.RunContainer(ctx, workbookRoot, workbookID, spec, time.Now, invocation.Options{ScopeResult: string(plan.Scope.Result), ScopeTarget: plan.Target, OnStarted: func(id string) { _ = launchInvocationWorkspace(workbookName, plan.Playbook, id) }})
		if err != nil {
			return image, fmt.Errorf("invocation %s failed: %w", record.ID, err)
		}
		if action == tuipkg.PortDiscovery {
			portsPath := filepath.Join(workbookRoot, "artifacts", "derived", record.ID, "ports.jsonl")
			if _, _, err := playbookpkg.IngestNaabuJournal(portsPath, record.ID, currentScope, log, time.Now); err != nil {
				return image, err
			}
		}
		return image, nil
	}
	model := tuipkg.NewWorkbenchModel(prefix, workbookName, addScope, refresh, journal.NewTail(log.Path()))
	model.SetScanRunner(runScan)
	model.SetDiscoveryLoader(func() ([]workbookview.Discovery, error) { return workbookview.Discoveries(root, workbookName) })
	program := tea.NewProgram(
		model,
		tea.WithInput(os.Stdin),
		tea.WithOutput(stdout),
		tea.WithEnvironment(os.Environ()),
	)
	result, err := program.Run()
	if err != nil {
		return "", err
	}
	model, ok := result.(*tuipkg.WorkbenchModel)
	if !ok || model.Selected() == 0 {
		return "", fmt.Errorf("action selection ended without a choice")
	}
	return strings.ToLower(string(model.Selected())), nil
}

func runBar(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] != "once" && args[0] != "stream" {
		fmt.Fprintln(stderr, "usage: akilix bar <once|stream>")
		return 2
	}
	render := func() []statusbar.Block {
		var blocks []statusbar.Block
		m, err := statusbar.Current(os.Getenv("XDG_RUNTIME_DIR"))
		if err != nil {
			blocks = []statusbar.Block{{FullText: " Akilix · no active workbook ", Color: "#aab1af", Separator: false}}
		} else {
			blocks = statusbar.Blocks(m)
		}
		return append(blocks, statusbar.TimeBlocks(time.Now())...)
	}
	if args[0] == "once" {
		b, _ := json.Marshal(render())
		fmt.Fprintln(stdout, string(b))
		return 0
	}
	stream := bufio.NewWriter(stdout)
	if _, err := fmt.Fprintln(stream, `{"version":1}`); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if _, err := fmt.Fprintln(stream, "["); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	for {
		b, err := json.Marshal(render())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if _, err := fmt.Fprintf(stream, "%s,\n", b); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if err := stream.Flush(); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		time.Sleep(time.Second)
	}
}

func launchWorkbookLog(name string) error {
	terminal, err := exec.LookPath("foot")
	if err != nil {
		return fmt.Errorf("open workbook log window: %w", err)
	}
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate akilix executable: %w", err)
	}
	cmd := exec.Command(terminal, "--title", "Akilix · "+name+" · Activity", self, "workbook", "follow", name)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open workbook log window: %w", err)
	}
	return cmd.Process.Release()
}

func launchInvocationWorkspace(workbookName, playbook, id string) error {
	if os.Getenv("SWAYSOCK") == "" {
		return nil
	}
	swaymsg, err := exec.LookPath("swaymsg")
	if err != nil {
		return err
	}
	terminal, err := exec.LookPath("foot")
	if err != nil {
		return err
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	tool := "run"
	if strings.Contains(playbook, "port") {
		tool = "naabu"
	} else if strings.Contains(playbook, "network") {
		tool = "nmap"
	}
	short := id
	if len(short) > 4 {
		short = short[len(short)-4:]
	}
	appID, workspace := "akilix-invocation-"+id, tool+"-"+short
	criteria := `[app_id="` + appID + `"]`
	command := `move container to workspace "` + workspace + `"`
	if err := exec.Command(swaymsg, "for_window", criteria, command).Run(); err != nil {
		return fmt.Errorf("register invocation workspace: %w", err)
	}
	cmd := exec.Command(terminal, "--app-id", appID, "--title", "Akilix · "+workspace, self, "workbook", "follow", workbookName, "--invocation", id)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func runRepository(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 || args[0] != "list" && args[0] != "show" {
		fmt.Fprintln(stderr, "usage: akilix repository list [--json] | repository show ID [--json]")
		return 2
	}
	path := os.Getenv("AKILIX_REPOSITORY_MANIFEST")
	if path == "" {
		path = "/usr/share/akilix/repositories.json"
		if _, err := os.Stat("repositories/repositories.json"); err == nil {
			path = "repositories/repositories.json"
		}
	}
	set, err := repositorypkg.Load(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if args[0] == "list" {
		if len(args) != 1 && !(len(args) == 2 && args[1] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix repository list [--json]")
			return 2
		}
		if len(args) == 2 {
			data, err := json.MarshalIndent(set.Repositories, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, item := range set.Repositories {
			fmt.Fprintf(stdout, "%s\t%s\t%s\tenabled=%t\n", item.ID, item.Status, item.Tier, item.ImageEnabled)
		}
		return 0
	}
	if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
		fmt.Fprintln(stderr, "usage: akilix repository show ID [--json]")
		return 2
	}
	item, err := set.Find(args[1])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(args) == 3 {
		data, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	fmt.Fprintf(stdout, "%s\nID: %s\nPurpose: %s\nTier: %s\nStatus: %s\nImage enabled: %t\nURL: %s\nSigning key: %s\n", item.Name, item.ID, item.Purpose, item.Tier, item.Status, item.ImageEnabled, item.BaseURL, item.KeyFingerprint)
	return 0
}

func runLogging(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") || len(args) > 0 && args[0] != "status" {
		fmt.Fprintln(stderr, "usage: akilix logging status WORKBOOK [--json]")
		return 2
	}
	root := effectiveWorkbookRoot()
	if _, err := workbook.Open(root, args[1]); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	policy, err := logpolicy.Load(filepath.Join(root, args[1]))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(args) == 3 {
		data, err := json.MarshalIndent(policy, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	fmt.Fprintf(stdout, "Command metadata         %s\nCommand arguments        %s\nContainer metadata       %s\nEvidence hashing         %s\nstdout capture           %s\nstderr capture           %s\nGenerated-file tracking  %s\nPacket metadata          %s\nTerminal recording       %s\n",
		enabled(policy.CommandMetadata), enabled(policy.CommandArguments), enabled(policy.ContainerMetadata),
		enabled(policy.EvidenceHashing), enabled(policy.StdoutCapture), enabled(policy.StderrCapture),
		enabled(policy.GeneratedFileTracking), enabled(policy.PacketMetadata), enabled(policy.TerminalRecording))
	return 0
}

func enabled(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}

func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 && !(len(args) == 2 && args[1] == "--json") || args[0] != "show" && args[0] != "path" {
		fmt.Fprintln(stderr, "usage: akilix config show [--json] | config path")
		return 2
	}
	settings, err := config.Effective(os.Getenv, func(path string) error { _, err := os.Stat(path); return err })
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if args[0] == "path" {
		fmt.Fprintln(stdout, settings.ConfigFile)
		return 0
	}
	if len(args) == 2 {
		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	fmt.Fprintf(stdout, "Config: %s\nState: %s\nWorkbooks: %s\nProfiles: %s\n", settings.ConfigFile, settings.StateDir, settings.WorkbookRoot, settings.ProfileDir)
	return 0
}

func runProfile(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 || (args[0] != "list" && args[0] != "show" && args[0] != "plan" && args[0] != "preflight" && args[0] != "verify" && args[0] != "history") {
		fmt.Fprintln(stderr, "usage: akilix profile list [--json] | profile show ID [--json] | profile plan ID [--json] | profile preflight ID [--json] | profile verify ID [--json] | profile history [ID] [--json]")
		return 2
	}
	if args[0] == "history" {
		if len(args) > 3 || len(args) == 3 && args[2] != "--json" {
			fmt.Fprintln(stderr, "usage: akilix profile history [ID] [--json]")
			return 2
		}
		profileID := ""
		jsonOutput := false
		if len(args) >= 2 {
			if args[1] == "--json" {
				jsonOutput = true
			} else {
				profileID = args[1]
			}
		}
		if len(args) == 3 {
			jsonOutput = true
		}
		history, err := profilepkg.VerificationHistory(effectiveStateDir(), profileID)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if jsonOutput {
			data, marshalErr := json.MarshalIndent(history, "", "  ")
			if marshalErr != nil {
				fmt.Fprintln(stderr, marshalErr)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, record := range history {
			state := "MISSING"
			if record.Ready {
				state = "READY"
			}
			fmt.Fprintf(stdout, "%s\t%s\t%-7s\t%s\n", record.VerifiedAt.Format(time.RFC3339), record.ProfileID, state, record.ID)
		}
		return 0
	}
	dir := os.Getenv("AKILIX_PROFILE_DIR")
	if dir == "" {
		if _, err := os.Stat("profiles"); err == nil {
			dir = "profiles"
		} else {
			settings, err := config.Effective(os.Getenv, func(path string) error { _, err := os.Stat(path); return err })
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			dir = settings.ProfileDir
		}
	}
	if args[0] == "list" {
		if len(args) != 1 && !(len(args) == 2 && args[1] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix profile list [--json]")
			return 2
		}
		items, err := profilepkg.LoadDir(dir)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 2 {
			data, err := json.MarshalIndent(items, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, item := range items {
			fmt.Fprintf(stdout, "%s\t%s\t%s\n", item.ID, item.Status, item.Name)
		}
		return 0
	}
	if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
		fmt.Fprintln(stderr, "usage: akilix profile show ID [--json]")
		return 2
	}
	item, err := profilepkg.Find(dir, args[1])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if args[0] == "plan" {
		plan := profilepkg.BuildPlan(item)
		if len(args) == 3 {
			data, err := json.MarshalIndent(plan, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, step := range plan.Steps {
			fmt.Fprintf(stdout, "%s\t%s\n", step.Phase, step.Action)
		}
		return 0
	}
	if args[0] == "preflight" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, err := profilepkg.PreflightHost(ctx, item, profilepkg.ExecInspector{})
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 3 {
			data, marshalErr := json.MarshalIndent(result, "", "  ")
			if marshalErr != nil {
				fmt.Fprintln(stderr, marshalErr)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
		} else {
			for _, check := range result.Checks {
				state := "BLOCKED"
				if check.Ready {
					state = "READY"
				}
				fmt.Fprintf(stdout, "%-22s %-7s %s\n", check.Name, state, check.Detail)
			}
			fmt.Fprintf(stdout, "apply supported: %t\nprivilege required for future apply: %t\n", result.ApplySupported, result.RequiresPrivilege)
		}
		if !result.Ready {
			return 1
		}
		return 0
	}
	if args[0] == "verify" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		verification, err := profilepkg.Verify(ctx, item, profilepkg.ExecRunner{}, time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		path, err := profilepkg.RecordVerification(effectiveStateDir(), verification)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 3 {
			data, marshalErr := json.MarshalIndent(verification, "", "  ")
			if marshalErr != nil {
				fmt.Fprintln(stderr, marshalErr)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
		} else {
			for _, component := range verification.Components {
				state := "MISSING"
				if component.Present {
					state = "READY"
				}
				fmt.Fprintf(stdout, "%s\t%-7s\t%s\n", component.Kind, state, component.Name)
			}
			fmt.Fprintf(stdout, "record: %s\n", path)
		}
		if !verification.Ready {
			return 1
		}
		return 0
	}
	if len(args) == 3 {
		data, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	fmt.Fprintf(stdout, "%s\nID: %s\nStatus: %s\n%s\n", item.Name, item.ID, item.Status, item.Description)
	return 0
}

func runCompletion(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: akilix completion <zsh|bash>")
		return 2
	}
	switch args[0] {
	case "zsh":
		fmt.Fprint(stdout, completion.Zsh)
		return 0
	case "bash":
		fmt.Fprint(stdout, completion.Bash)
		return 0
	default:
		fmt.Fprintln(stderr, "unsupported shell")
		return 2
	}
}

func runContainer(args []string, stdout, stderr io.Writer) int {
	if len(args) >= 1 && args[0] == "doctor" {
		if len(args) > 2 || len(args) == 2 && args[1] != "--json" {
			fmt.Fprintln(stderr, "usage: akilix container doctor [--json]")
			return 2
		}
		status, err := containerpkg.CheckRuntime(context.Background(), containerpkg.PodmanRunner{})
		if len(args) == 2 {
			data, marshalErr := json.MarshalIndent(status, "", "  ")
			if marshalErr != nil {
				fmt.Fprintln(stderr, marshalErr)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
		} else {
			fmt.Fprintf(stdout, "Podman available: %t\nRootless: %t\nUser namespace ready: %t\n", status.Available, status.Rootless, status.UserNamespaceReady)
		}
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	if len(args) >= 1 && args[0] == "run" {
		return runContainerCommand(args[1:], stdout, stderr)
	}
	if len(args) != 2 || args[0] != "inspect" {
		fmt.Fprintln(stderr, "usage: akilix container doctor [--json] | container inspect IMAGE | container run WORKBOOK IMAGE [--mount-originals] [--target TARGET] [--override] [--json] [--workdir DIR] [--env KEY=VALUE] -- COMMAND [ARGS...]")
		return 2
	}
	id, err := containerpkg.Resolve(context.Background(), containerpkg.PodmanRunner{}, args[1])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	data, err := json.MarshalIndent(id, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, string(data))
	return 0
}

func runContainerCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) < 4 {
		fmt.Fprintln(stderr, "usage: akilix container run WORKBOOK IMAGE [--mount-originals] [--target TARGET] [--override] [--json] [--workdir DIR] [--env KEY=VALUE] -- COMMAND [ARGS...]")
		return 2
	}
	target, override, workdir := "", false, ""
	environment := map[string]string{}
	jsonOutput := false
	mountOriginals, mountOutput := false, false
	separator := -1
	for i := 2; i < len(args); i++ {
		if args[i] == "--target" && i+1 < len(args) {
			target = args[i+1]
			i++
			continue
		}
		if args[i] == "--json" {
			jsonOutput = true
			continue
		}
		if args[i] == "--mount-originals" {
			if mountOriginals {
				fmt.Fprintln(stderr, "duplicate --mount-originals")
				return 2
			}
			mountOriginals = true
			continue
		}
		if args[i] == "--mount-output" {
			if mountOutput {
				fmt.Fprintln(stderr, "duplicate --mount-output")
				return 2
			}
			mountOutput = true
			continue
		}
		if args[i] == "--workdir" && i+1 < len(args) {
			workdir = args[i+1]
			i++
			continue
		}
		if args[i] == "--env" && i+1 < len(args) {
			pair := args[i+1]
			i++
			key, value, found := strings.Cut(pair, "=")
			if !found || key == "" {
				fmt.Fprintln(stderr, "--env requires KEY=VALUE")
				return 2
			}
			if _, exists := environment[key]; exists {
				fmt.Fprintf(stderr, "duplicate --env key %s\n", key)
				return 2
			}
			environment[key] = value
			continue
		}
		if args[i] == "--override" {
			override = true
			continue
		}
		if args[i] == "--" {
			separator = i
			break
		}
		fmt.Fprintf(stderr, "invalid container option %s\n", args[i])
		return 2
	}
	if separator < 0 || separator+1 >= len(args) {
		fmt.Fprintln(stderr, "command is required after --")
		return 2
	}
	if override && target == "" {
		fmt.Fprintln(stderr, "--override requires --target")
		return 2
	}
	root := effectiveWorkbookRoot()
	m, err := workbook.Open(root, args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if m.Status != "open" {
		fmt.Fprintln(stderr, "workbook is closed")
		return 1
	}
	scopeResult := ""
	if target != "" {
		c, scopeErr := scope.Load(filepath.Join(root, args[0]))
		if scopeErr != nil {
			fmt.Fprintln(stderr, scopeErr)
			return 1
		}
		scopeResult = string(scope.Evaluate(c, target))
		if scopeResult == string(scope.Deny) && !override {
			fmt.Fprintf(stderr, "scope denied target %s; use --override to record an explicit override\n", target)
			return 1
		}
		if override && scopeResult != string(scope.Deny) {
			fmt.Fprintln(stderr, "--override is only valid for a denied target")
			return 2
		}
	}
	identity, err := containerpkg.Resolve(context.Background(), containerpkg.PodmanRunner{}, args[1])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	mounts := []containerpkg.Mount{}
	if mountOriginals {
		mount, mountErr := containerpkg.OriginalEvidenceMount(filepath.Join(root, args[0]))
		if mountErr != nil {
			fmt.Fprintln(stderr, mountErr)
			return 1
		}
		mounts = append(mounts, mount)
	}
	spec := containerpkg.Spec{Identity: identity, Arguments: append([]string(nil), args[separator+1:]...), Mounts: mounts, InvocationOutput: mountOutput, Workdir: workdir, Environment: environment}
	r, err := invocation.RunContainer(context.Background(), filepath.Join(root, args[0]), m.ID, spec, time.Now, invocation.Options{ScopeResult: scopeResult, ScopeTarget: target, ScopeOverride: override})
	if err != nil {
		fmt.Fprintf(stderr, "invocation %s failed: %v\n", r.ID, err)
		return r.ExitCode
	}
	if jsonOutput {
		data, marshalErr := json.MarshalIndent(r, "", "  ")
		if marshalErr != nil {
			fmt.Fprintln(stderr, marshalErr)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	fmt.Fprintf(stdout, "invocation %s complete\n", r.ID)
	return 0
}

func runCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) >= 1 && args[0] == "list" {
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix run list WORKBOOK")
			return 2
		}
		root := effectiveWorkbookRoot()
		if _, err := workbook.Open(root, args[1]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		records, err := invocation.List(filepath.Join(root, args[1]))
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 3 {
			data, err := json.MarshalIndent(records, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, r := range records {
			fmt.Fprintf(stdout, "%s\t%s\t%d\t%s\n", r.ID, r.Status, r.ExitCode, strings.Join(r.Arguments, " "))
		}
		return 0
	}
	if len(args) < 3 {
		fmt.Fprintln(stderr, "usage: akilix run WORKBOOK [--target TARGET] [--override] -- COMMAND [ARGS...]")
		return 2
	}
	workbookName := args[0]
	target := ""
	override := false
	separator := -1
	for i := 1; i < len(args); i++ {
		if args[i] == "--" {
			separator = i
			break
		}
		if args[i] == "--target" && i+1 < len(args) {
			target = args[i+1]
			i++
			continue
		}
		if args[i] == "--override" {
			override = true
			continue
		}
		fmt.Fprintln(stderr, "invalid run option")
		return 2
	}
	if separator < 0 || separator+1 >= len(args) {
		fmt.Fprintln(stderr, "command is required after --")
		return 2
	}
	if override && target == "" {
		fmt.Fprintln(stderr, "--override requires --target")
		return 2
	}
	root := effectiveWorkbookRoot()
	wbRoot := filepath.Join(root, workbookName)
	m, err := workbook.Open(root, workbookName)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if m.Status != "open" {
		fmt.Fprintln(stderr, "workbook is closed")
		return 1
	}
	scopeResult := ""
	if target != "" {
		c, scopeErr := scope.Load(wbRoot)
		if scopeErr != nil {
			fmt.Fprintln(stderr, scopeErr)
			return 1
		}
		scopeResult = string(scope.Evaluate(c, target))
		if scopeResult == string(scope.Deny) && !override {
			fmt.Fprintf(stderr, "scope denied target %s; use --override to record an explicit override\n", target)
			return 1
		}
		if override && scopeResult != string(scope.Deny) {
			fmt.Fprintln(stderr, "--override is only valid for a denied target")
			return 2
		}
	}
	r, err := invocation.RunWithOptions(context.Background(), wbRoot, m.ID, args[separator+1:], time.Now, invocation.Options{ScopeResult: scopeResult, ScopeTarget: target, ScopeOverride: override})
	if err != nil {
		fmt.Fprintf(stderr, "invocation %s failed: %v\n", r.ID, err)
		return r.ExitCode
	}
	fmt.Fprintf(stdout, "invocation %s complete\n", r.ID)
	return 0
}

func runScope(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: akilix scope <add|remove|exclude|list|check> WORKBOOK [TARGET]")
		return 2
	}
	root := effectiveWorkbookRoot()
	workbookRoot := filepath.Join(root, args[1])
	m, err := workbook.Open(root, args[1])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	c, err := scope.Load(workbookRoot)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	switch args[0] {
	case "add", "exclude":
		if m.Status != "open" {
			fmt.Fprintln(stderr, "workbook is closed")
			return 1
		}
		if len(args) != 3 {
			fmt.Fprintln(stderr, "target is required")
			return 2
		}
		log, err := journal.Open(workbookRoot)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if _, err := scope.AddRecorded(workbookRoot, m.ID, args[2], args[0] == "exclude", log, time.Now().UTC()); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "remove":
		if m.Status != "open" {
			fmt.Fprintln(stderr, "workbook is closed")
			return 1
		}
		if len(args) != 3 {
			fmt.Fprintln(stderr, "target is required")
			return 2
		}
		log, err := journal.Open(workbookRoot)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if _, err := scope.RemoveRecorded(workbookRoot, m.ID, args[2], log, time.Now().UTC()); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "list":
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
			return 2
		}
		if len(args) == 3 {
			data, err := json.MarshalIndent(map[string][]string{"include": c.Includes, "exclude": c.Excludes}, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, v := range c.Includes {
			fmt.Fprintf(stdout, "ALLOW\t%s\n", v)
		}
		for _, v := range c.Excludes {
			fmt.Fprintf(stdout, "DENY\t%s\n", v)
		}
		return 0
	case "check":
		if len(args) != 3 && !(len(args) == 4 && args[3] == "--json") {
			fmt.Fprintln(stderr, "target is required")
			return 2
		}
		decision := scope.EvaluateDecision(c, args[2])
		if len(args) == 4 {
			data, err := json.Marshal(decision)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		fmt.Fprintln(stdout, decision.Result)
		return 0
	default:
		fmt.Fprintln(stderr, "unknown scope command")
		return 2
	}
}

func runEvidence(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: akilix evidence <import|list|verify> WORKBOOK [SOURCE|ID]")
		return 2
	}
	if len(args) < 2 {
		fmt.Fprintln(stderr, "workbook name is required")
		return 2
	}
	root := effectiveWorkbookRoot()
	workbookRoot := filepath.Join(root, args[1])
	m, err := workbook.Open(root, args[1])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if m.Status != "open" && args[0] == "import" {
		fmt.Fprintln(stderr, "workbook is closed")
		return 1
	}
	switch args[0] {
	case "import":
		if len(args) != 3 {
			fmt.Fprintln(stderr, "usage: akilix evidence import WORKBOOK SOURCE")
			return 2
		}
		r, err := evidence.ImportRecorded(workbookRoot, m.ID, args[2], time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "imported %s (%s) sha256=%s\n", r.Filename, r.ID, r.SHA256)
		return 0
	case "list":
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix evidence list WORKBOOK")
			return 2
		}
		records, err := evidence.List(workbookRoot)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 3 {
			data, err := json.MarshalIndent(records, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, r := range records {
			verification := r.Verification
			if verification == "" {
				verification = "unverified"
			}
			fmt.Fprintf(stdout, "%s\t%s\t%d\t%s\t%s\n", r.ID, r.Filename, r.Size, verification, r.SHA256)
		}
		return 0
	case "verify":
		if len(args) == 3 && args[2] == "--all" || len(args) == 4 && args[2] == "--all" && args[3] == "--json" {
			records, allMatch, err := evidence.VerifyAllRecorded(workbookRoot, m.ID, time.Now)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			if len(args) == 4 {
				data, marshalErr := json.MarshalIndent(records, "", "  ")
				if marshalErr != nil {
					fmt.Fprintln(stderr, marshalErr)
					return 1
				}
				fmt.Fprintln(stdout, string(data))
				if !allMatch {
					return 1
				}
				return 0
			}
			for _, record := range records {
				fmt.Fprintf(stdout, "%s\t%s\n", record.ID, strings.ToUpper(record.Verification))
			}
			if !allMatch {
				return 1
			}
			return 0
		}
		if len(args) != 3 && !(len(args) == 4 && args[3] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix evidence verify WORKBOOK EVIDENCE_ID [--json] | --all [--json]")
			return 2
		}
		ok, record, err := evidence.VerifyRecorded(workbookRoot, m.ID, args[2], time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 4 {
			data, marshalErr := json.Marshal(map[string]string{"evidence_id": record.ID, "verification": record.Verification})
			if marshalErr != nil {
				fmt.Fprintln(stderr, marshalErr)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			if !ok {
				return 1
			}
			return 0
		}
		if !ok {
			fmt.Fprintln(stdout, "MISMATCH")
			return 1
		}
		fmt.Fprintln(stdout, "MATCH")
		return 0
	default:
		fmt.Fprintln(stderr, "unknown evidence command")
		return 2
	}
}

func runWorkbook(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: akilix workbook <create|list|open|activate|active|deactivate|overview|discoveries|follow|path|status|close|reopen|rename|validate>")
		return 2
	}
	root := effectiveWorkbookRoot()
	switch args[0] {
	case "activate":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: akilix workbook activate NAME")
			return 2
		}
		if err := statusbar.Activate(os.Getenv("XDG_RUNTIME_DIR"), root, args[1]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "active workbook: %s\n", args[1])
		return 0
	case "active":
		if len(args) != 1 && !(len(args) == 2 && args[1] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix workbook active [--json]")
			return 2
		}
		state, err := statusbar.Active(os.Getenv("XDG_RUNTIME_DIR"))
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 2 {
			data, err := json.Marshal(state)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		fmt.Fprintf(stdout, "%s\t%s\n", state.Workbook, state.Root)
		return 0
	case "deactivate":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "usage: akilix workbook deactivate")
			return 2
		}
		if err := statusbar.Deactivate(os.Getenv("XDG_RUNTIME_DIR")); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, "workbook context deactivated")
		return 0
	case "discoveries":
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix workbook discoveries NAME [--json]")
			return 2
		}
		items, err := workbookview.Discoveries(root, args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 3 {
			data, err := json.MarshalIndent(items, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, item := range items {
			label := item.Value
			if item.Hostname != "" {
				label += " (" + item.Hostname + ")"
			}
			fmt.Fprintf(stdout, "%-4s  %-36s seen=%d  last=%s  %s\n", item.Kind, label, item.Occurrences, item.LastSeen, item.LastProvenanceID)
		}
		return 0
	case "follow":
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--once") && !(len(args) == 4 && args[2] == "--invocation" && args[3] != "") {
			fmt.Fprintln(stderr, "usage: akilix workbook follow NAME [--once | --invocation ID]")
			return 2
		}
		if _, err := workbook.Open(root, args[1]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "Akilix workbook activity · %s\nWaiting for canonical invocation records…\n\n", args[1])
		seen := 0
		var stdoutOffset, stderrOffset int64
		for {
			events, err := activity.List(filepath.Join(root, args[1]))
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			for _, event := range events[seen:] {
				if len(args) == 4 && event.InvocationID != args[3] {
					continue
				}
				exit := ""
				if event.ExitCode != nil {
					exit = fmt.Sprintf(" exit=%d", *event.ExitCode)
				}
				fmt.Fprintf(stdout, "%s  %-9s %-9s %-18s%s  %s\n", event.Timestamp.Local().Format("15:04:05"), event.Phase, event.Executor, event.Tool, exit, event.InvocationID)
			}
			seen = len(events)
			if len(args) == 4 {
				stdoutOffset = copyGrowingFile(filepath.Join(root, args[1], "tool-output", args[3]+".stdout"), stdoutOffset, stdout)
				stderrOffset = copyGrowingFile(filepath.Join(root, args[1], "tool-output", args[3]+".stderr"), stderrOffset, stderr)
			}
			if len(args) == 3 {
				return 0
			}
			time.Sleep(750 * time.Millisecond)
		}
	case "overview":
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix workbook overview NAME [--json]")
			return 2
		}
		overview, err := workbookview.Build(root, args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 3 {
			data, err := json.MarshalIndent(overview, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		fmt.Fprint(stdout, workbookview.Render(overview))
		return 0
	case "path":
		if len(args) != 2 && len(args) != 3 {
			fmt.Fprintln(stderr, "usage: akilix workbook path NAME [SECTION]")
			return 2
		}
		section := "root"
		if len(args) == 3 {
			section = args[2]
		}
		path, err := workbookview.Path(root, args[1], section)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, path)
		return 0
	case "validate":
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix workbook validate NAME [--json]")
			return 2
		}
		workbookRoot := filepath.Join(root, args[1])
		if err := workbook.ValidateLayout(root, args[1]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if _, err := scope.Load(workbookRoot); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if _, err := logpolicy.Load(workbookRoot); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		evidenceRecords, allEvidenceMatch, err := evidence.CheckAll(workbookRoot)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		invocationRecords, err := invocation.List(workbookRoot)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if !allEvidenceMatch {
			fmt.Fprintln(stderr, "evidence integrity mismatch")
			return 1
		}
		if len(args) == 3 {
			data, err := json.Marshal(map[string]interface{}{"valid": true, "workbook": args[1], "evidence": len(evidenceRecords), "invocations": len(invocationRecords)})
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		fmt.Fprintf(stdout, "valid %s\n", args[1])
		return 0
	case "create":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: akilix workbook create NAME")
			return 2
		}
		m, err := workbook.Create(root, args[1], timeNow())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "created %s (%s)\n", m.Name, m.ID)
		return 0
	case "rename":
		if len(args) != 3 {
			fmt.Fprintln(stderr, "usage: akilix workbook rename OLD_NAME NEW_NAME")
			return 2
		}
		m, err := workbook.Rename(root, args[1], args[2])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "renamed %s (%s)\n", m.Name, m.ID)
		return 0
	case "list":
		if len(args) != 1 && !(len(args) == 2 && args[1] == "--json") {
			fmt.Fprintln(stderr, "usage: akilix workbook list")
			return 2
		}
		items, err := workbook.List(root)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(args) == 2 {
			data, err := json.MarshalIndent(items, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		for _, m := range items {
			fmt.Fprintf(stdout, "%s\t%s\t%s\n", m.Name, m.ID, m.Status)
		}
		return 0
	case "open", "status", "close", "reopen":
		if len(args) != 2 && !(args[0] == "status" && len(args) == 3 && args[2] == "--json") {
			fmt.Fprintf(stderr, "usage: akilix workbook %s NAME\n", args[0])
			return 2
		}
		if args[0] == "close" {
			m, err := workbook.SetStatus(root, args[1], "closed")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintf(stdout, "closed %s\n", m.Name)
			return 0
		}
		if args[0] == "reopen" {
			m, err := workbook.SetStatus(root, args[1], "open")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintf(stdout, "reopened %s\n", m.Name)
			return 0
		}
		m, err := workbook.Open(root, args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if args[0] == "status" && len(args) == 3 {
			data, err := json.MarshalIndent(m, "", "  ")
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprintln(stdout, string(data))
			return 0
		}
		if args[0] == "open" {
			overview, err := workbookview.Build(root, args[1])
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprint(stdout, workbookview.Render(overview))
			return 0
		}
		fmt.Fprintf(stdout, "%s\nID: %s\nStatus: %s\nCreated: %s\n", m.Name, m.ID, m.Status, m.Created.Format(time.RFC3339Nano))
		return 0
	default:
		fmt.Fprintln(stderr, "unknown workbook command")
		return 2
	}
}

func copyGrowingFile(path string, offset int64, output io.Writer) int64 {
	f, err := os.Open(path)
	if err != nil {
		return offset
	}
	defer f.Close()
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return offset
	}
	n, err := io.Copy(output, f)
	if err != nil {
		return offset
	}
	return offset + n
}

func effectiveWorkbookRoot() string {
	settings, err := config.Effective(os.Getenv, func(path string) error { _, err := os.Stat(path); return err })
	if err == nil {
		return settings.WorkbookRoot
	}
	_, state := config.UserPaths()
	return filepath.Join(state, "workbooks")
}

func effectiveStateDir() string {
	settings, err := config.Effective(os.Getenv, func(path string) error { _, err := os.Stat(path); return err })
	if err == nil {
		return settings.StateDir
	}
	_, state := config.UserPaths()
	return state
}

func usbIDsPath() string {
	if path := os.Getenv("AKILIX_USB_IDS_PATH"); path != "" {
		return path
	}
	return "/usr/share/hwdata/usb.ids"
}

var timeNow = func() time.Time { return time.Now().UTC() }
