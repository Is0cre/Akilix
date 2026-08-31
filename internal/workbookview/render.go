package workbookview

import (
	"fmt"
	"strings"
)

// Render returns a stable, terminal-oriented workbook workspace. It deliberately
// contains no terminal control sequences so redirected output remains readable.
func Render(overview Overview) string {
	var b strings.Builder
	writeRule := func(title string) {
		fmt.Fprintf(&b, "├─ %s %s\n", title, strings.Repeat("─", max(1, 70-len(title))))
	}

	fmt.Fprintln(&b, "╭─ PenSUSE / WORKBOOK ────────────────────────────────────────────────────")
	fmt.Fprintf(&b, "│  %s  [%s]\n", overview.Name, strings.ToUpper(overview.Status))
	writeRule("CASE")
	fmt.Fprintf(&b, "│  ID       %s\n", overview.ID)
	fmt.Fprintf(&b, "│  Created  %s\n", overview.Created)
	fmt.Fprintf(&b, "│  Root     %s\n", overview.Root)
	writeRule("WORKSPACE")
	fmt.Fprintf(&b, "│  Scope         %d included · %d excluded\n", overview.ScopeIncludes, overview.ScopeExcludes)
	fmt.Fprintf(&b, "│  Evidence      %d originals\n", overview.Evidence)
	fmt.Fprintf(&b, "│  Invocations   %d total · %d failed\n", overview.Invocations, overview.FailedInvocations)
	writeRule("CAPTURE POLICY")
	fmt.Fprintf(&b, "│  stdout %-8s  stderr %-8s  terminal %-8s  packet metadata %s\n",
		state(overview.Logging.StdoutCapture), state(overview.Logging.StderrCapture),
		state(overview.Logging.TerminalRecording), state(overview.Logging.PacketMetadata))
	writeRule("NAVIGATE")
	for _, section := range overview.Sections {
		fmt.Fprintf(&b, "│  %-18s %s\n", section.Name, section.Path)
	}
	writeRule("QUICK JUMPS")
	fmt.Fprintf(&b, "│  cd \"$(pensuse workbook path %s evidence)\"\n", overview.Name)
	fmt.Fprintf(&b, "│  cd \"$(pensuse workbook path %s reports)\"\n", overview.Name)
	fmt.Fprintln(&b, "╰─ Read-only view · no network activity · original evidence stays immutable")
	return b.String()
}

func state(value bool) string {
	if value {
		return "ON"
	}
	return "OFF"
}
