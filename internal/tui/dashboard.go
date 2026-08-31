package tui

import (
	"fmt"
	"net"
	"strings"

	"github.com/pensuse/pensuse/internal/scope"
	"github.com/pensuse/pensuse/internal/workbookview"
)

type Palette struct{ Blue, Green, Amber, Red, Purple, Muted, Reset string }

func ANSI(enabled bool) Palette {
	if !enabled {
		return Palette{}
	}
	return Palette{Blue: "\x1b[94m", Green: "\x1b[92m", Amber: "\x1b[93m", Red: "\x1b[91m", Purple: "\x1b[95m", Muted: "\x1b[90m", Reset: "\x1b[0m"}
}

func Render(overview workbookview.Overview, config scope.Config, color bool) string {
	p := ANSI(color)
	stateColor := p.Green
	if overview.Status != "open" {
		stateColor = p.Amber
	}
	healthColor, health := p.Green, "healthy"
	if overview.FailedInvocations > 0 {
		healthColor, health = p.Red, "attention"
	}
	cidrs := 0
	for _, value := range config.Includes {
		if _, _, err := net.ParseCIDR(value); err == nil {
			cidrs++
		}
	}
	readyColor, ready := p.Green, "READY"
	if overview.Status != "open" || cidrs == 0 {
		readyColor, ready = p.Amber, "NEEDS SCOPE"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s┌─ PenSUSE ─ Workbook workspace ──────────────────────────────┐%s\n", p.Blue, p.Reset)
	b.WriteString(frameRow(fmt.Sprintf("%s%s%s  %s%s%s  ID %.18s…", p.Blue, overview.Name, p.Reset, stateColor, strings.ToUpper(overview.Status), p.Reset, overview.ID)))
	fmt.Fprintf(&b, "├─ Scope ───────────────────┬─ Activity ──────────────────────┤\n")
	b.WriteString(frameRow(fmt.Sprintf("Includes %-3d  CIDRs %-3d  │  Invocations %-4d  %s%s%s", overview.ScopeIncludes, cidrs, overview.Invocations, healthColor, health, p.Reset)))
	b.WriteString(frameRow(fmt.Sprintf("Excludes %-3d             │  Evidence %-4d  failures %-3d", overview.ScopeExcludes, overview.Evidence, overview.FailedInvocations)))
	fmt.Fprintf(&b, "├─ Playbooks ─────────────────────────────────────────────────┤\n")
	b.WriteString(frameRow(fmt.Sprintf("Local network discovery  %s%s%s", readyColor, ready, p.Reset)))
	b.WriteString(frameRow(fmt.Sprintf("Local port discovery     %s%s%s", readyColor, ready, p.Reset)))
	b.WriteString(frameRow(fmt.Sprintf("%sPlanning is passive. Execution always needs confirmation.%s", p.Muted, p.Reset)))
	fmt.Fprintf(&b, "%s└─────────────────────────────────────────────────────────────┘%s\n", p.Blue, p.Reset)
	return b.String()
}

func frameRow(content string) string {
	const width = 60
	padding := width - visibleWidth(content)
	if padding < 0 {
		padding = 0
	}
	return "│ " + content + strings.Repeat(" ", padding) + "│\n"
}

func visibleWidth(value string) int {
	width, escape := 0, false
	for _, r := range value {
		if r == '\x1b' {
			escape = true
			continue
		}
		if escape {
			if r == 'm' {
				escape = false
			}
			continue
		}
		width++
	}
	return width
}
