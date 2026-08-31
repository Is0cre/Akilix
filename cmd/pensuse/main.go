package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pensuse/pensuse/internal/completion"
	"github.com/pensuse/pensuse/internal/config"
	containerpkg "github.com/pensuse/pensuse/internal/container"
	"github.com/pensuse/pensuse/internal/evidence"
	"github.com/pensuse/pensuse/internal/invocation"
	"github.com/pensuse/pensuse/internal/logpolicy"
	profilepkg "github.com/pensuse/pensuse/internal/profile"
	repositorypkg "github.com/pensuse/pensuse/internal/repository"
	"github.com/pensuse/pensuse/internal/scope"
	"github.com/pensuse/pensuse/internal/version"
	"github.com/pensuse/pensuse/internal/workbook"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: pensuse version [--json] | workbook ... | scope ... | evidence ... | run ... | container ...")
		return 2
	}
	if args[0] == "workbook" {
		return runWorkbook(args[1:], stdout, stderr)
	}
	if args[0] == "evidence" {
		return runEvidence(args[1:], stdout, stderr)
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
	if args[0] != "version" {
		fmt.Fprintln(stderr, "usage: pensuse version [--json] | workbook ... | scope ... | evidence ... | run ... | container ...")
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
		fmt.Fprintln(stderr, "usage: pensuse version [--json]")
		return 2
	}
	fmt.Fprintf(stdout, "%s %s\nBase: %s\nArchitecture: %s\n", info.Name, info.Version, info.Base, info.Architecture)
	return 0
}

func runRepository(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 || args[0] != "list" && args[0] != "show" {
		fmt.Fprintln(stderr, "usage: pensuse repository list [--json] | repository show ID [--json]")
		return 2
	}
	path := os.Getenv("PENSUSE_REPOSITORY_MANIFEST")
	if path == "" {
		path = "/usr/share/pensuse/repositories.json"
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
			fmt.Fprintln(stderr, "usage: pensuse repository list [--json]")
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
		fmt.Fprintln(stderr, "usage: pensuse repository show ID [--json]")
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
		fmt.Fprintln(stderr, "usage: pensuse logging status WORKBOOK [--json]")
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
		fmt.Fprintln(stderr, "usage: pensuse config show [--json] | config path")
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
	if len(args) < 1 || (args[0] != "list" && args[0] != "show" && args[0] != "plan") {
		fmt.Fprintln(stderr, "usage: pensuse profile list [--json] | profile show ID [--json] | profile plan ID [--json]")
		return 2
	}
	dir := os.Getenv("PENSUSE_PROFILE_DIR")
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
			fmt.Fprintln(stderr, "usage: pensuse profile list [--json]")
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
		fmt.Fprintln(stderr, "usage: pensuse profile show ID [--json]")
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
		fmt.Fprintln(stderr, "usage: pensuse completion <zsh|bash>")
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
	if len(args) >= 1 && args[0] == "run" {
		return runContainerCommand(args[1:], stdout, stderr)
	}
	if len(args) != 2 || args[0] != "inspect" {
		fmt.Fprintln(stderr, "usage: pensuse container inspect IMAGE | container run WORKBOOK IMAGE [--target TARGET] [--override] [--json] [--workdir DIR] [--env KEY=VALUE] -- COMMAND [ARGS...]")
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
		fmt.Fprintln(stderr, "usage: pensuse container run WORKBOOK IMAGE [--target TARGET] [--override] [--json] [--workdir DIR] [--env KEY=VALUE] -- COMMAND [ARGS...]")
		return 2
	}
	target, override, workdir := "", false, ""
	environment := map[string]string{}
	jsonOutput := false
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
	spec := containerpkg.Spec{Identity: identity, Arguments: append([]string(nil), args[separator+1:]...), Workdir: workdir, Environment: environment}
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
			fmt.Fprintln(stderr, "usage: pensuse run list WORKBOOK")
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
		fmt.Fprintln(stderr, "usage: pensuse run WORKBOOK [--target TARGET] [--override] -- COMMAND [ARGS...]")
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
		fmt.Fprintln(stderr, "usage: pensuse scope <add|remove|exclude|list|check> WORKBOOK [TARGET]")
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
		if err := scope.Add(workbookRoot, args[2], args[0] == "exclude"); err != nil {
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
		if err := scope.Remove(workbookRoot, args[2], false); err != nil {
			if err = scope.Remove(workbookRoot, args[2], true); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
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
		fmt.Fprintln(stderr, "usage: pensuse evidence <import|list|verify> WORKBOOK [SOURCE|ID]")
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
			fmt.Fprintln(stderr, "usage: pensuse evidence import WORKBOOK SOURCE")
			return 2
		}
		r, err := evidence.Import(workbookRoot, args[2], time.Now().UTC())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "imported %s (%s) sha256=%s\n", r.Filename, r.ID, r.SHA256)
		return 0
	case "list":
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
			fmt.Fprintln(stderr, "usage: pensuse evidence list WORKBOOK")
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
			records, allMatch, err := evidence.VerifyAll(workbookRoot)
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
			fmt.Fprintln(stderr, "usage: pensuse evidence verify WORKBOOK EVIDENCE_ID [--json] | --all [--json]")
			return 2
		}
		ok, record, err := evidence.Verify(workbookRoot, args[2])
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
		fmt.Fprintln(stderr, "usage: pensuse workbook <create|list|open|status|close|reopen|rename|validate>")
		return 2
	}
	root := effectiveWorkbookRoot()
	switch args[0] {
	case "validate":
		if len(args) != 2 && !(len(args) == 3 && args[2] == "--json") {
			fmt.Fprintln(stderr, "usage: pensuse workbook validate NAME [--json]")
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
			fmt.Fprintln(stderr, "usage: pensuse workbook create NAME")
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
			fmt.Fprintln(stderr, "usage: pensuse workbook rename OLD_NAME NEW_NAME")
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
			fmt.Fprintln(stderr, "usage: pensuse workbook list")
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
			fmt.Fprintf(stderr, "usage: pensuse workbook %s NAME\n", args[0])
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
		fmt.Fprintf(stdout, "%s\nID: %s\nStatus: %s\nCreated: %s\n", m.Name, m.ID, m.Status, m.Created.Format(time.RFC3339Nano))
		return 0
	default:
		fmt.Fprintln(stderr, "unknown workbook command")
		return 2
	}
}

func effectiveWorkbookRoot() string {
	settings, err := config.Effective(os.Getenv, func(path string) error { _, err := os.Stat(path); return err })
	if err == nil {
		return settings.WorkbookRoot
	}
	_, state := config.UserPaths()
	return filepath.Join(state, "workbooks")
}

var timeNow = func() time.Time { return time.Now().UTC() }
