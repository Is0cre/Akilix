package greeter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

type Preflight struct {
	SELinux       string `json:"selinux"`
	Media         string `json:"media"`
	MediaWarning  bool   `json:"media_warning"`
	Containers    string `json:"containers"`
	WorkbookState string `json:"workbook_state"`
}

func Inspect(root string) Preflight {
	if root == "" {
		root = "/"
	}
	mediaCount, writable := blockMedia(root)
	return Preflight{
		SELinux:       selinuxState(root),
		Media:         formatMedia(mediaCount, writable),
		MediaWarning:  writable != 0,
		Containers:    containerState(root),
		WorkbookState: "Available after authentication",
	}
}

func selinuxState(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "sys/fs/selinux/enforce"))
	if err != nil {
		return "DISABLED / UNAVAILABLE"
	}
	if strings.TrimSpace(string(data)) == "1" {
		return "ENFORCING"
	}
	return "PERMISSIVE"
}

func blockMedia(root string) (count, writable int) {
	entries, err := os.ReadDir(filepath.Join(root, "sys/class/block"))
	if err != nil {
		return 0, 0
	}
	for _, entry := range entries {
		name := entry.Name()
		base := filepath.Join(root, "sys/class/block", name)
		if _, err := os.Stat(filepath.Join(base, "partition")); err == nil {
			continue
		}
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") || strings.HasPrefix(name, "zram") || strings.HasPrefix(name, "dm-") {
			continue
		}
		removable, _ := os.ReadFile(filepath.Join(base, "removable"))
		target, _ := filepath.EvalSymlinks(base)
		if strings.TrimSpace(string(removable)) != "1" && !strings.Contains(target, "/usb") {
			continue
		}
		count++
		readOnly, _ := os.ReadFile(filepath.Join(base, "ro"))
		if strings.TrimSpace(string(readOnly)) != "1" {
			writable++
		}
	}
	return count, writable
}

func formatMedia(count, writable int) string {
	if count == 0 {
		return "No peripheral media"
	}
	if writable == 0 {
		return fmt.Sprintf("%d peripheral media · READ-ONLY", count)
	}
	return fmt.Sprintf("%d peripheral media · %d NOT READ-ONLY", count, writable)
}

func containerState(root string) string {
	if _, err := os.Stat(filepath.Join(root, "usr/bin/podman")); err != nil {
		return "Podman unavailable"
	}
	data, err := os.ReadFile(filepath.Join(root, "etc/subuid"))
	if err != nil {
		return "Podman installed · rootless unverified"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "akilix:") {
			return "Podman · rootless configured"
		}
	}
	return "Podman installed · rootless unverified"
}

func Render(snapshot Preflight, color bool) string {
	accent := lipgloss.Color("#87ff00")
	warning := lipgloss.Color("#ff5f00")
	border := accent
	if snapshot.MediaWarning || snapshot.SELinux != "ENFORCING" {
		border = warning
	}
	brand := "     __\n _..'  `.._\n/  _    _  \\\n\\_/ \\__/ \\_/\n  `--o-o--'\n\nAKILIX\nSecurity work with provenance."
	rows := []string{
		"SYSTEM PRE-FLIGHT AUDIT",
		"",
		"SECURITY POLICY       [ SELinux: " + snapshot.SELinux + " ]",
		"LOCAL WORKBOOKS       [ " + snapshot.WorkbookState + " ]",
		"HARDWARE PROVENANCE   [ " + snapshot.Media + " ]",
		"CONTAINER RUNTIME     [ " + snapshot.Containers + " ]",
	}
	if color {
		brand = lipgloss.NewStyle().Foreground(accent).Render(brand)
		rows[0] = lipgloss.NewStyle().Foreground(accent).Bold(true).Render(rows[0])
		if snapshot.MediaWarning || snapshot.SELinux != "ENFORCING" {
			rows[2] = lipgloss.NewStyle().Foreground(warning).Render(rows[2])
			rows[4] = lipgloss.NewStyle().Foreground(warning).Render(rows[4])
		}
	}
	left := lipgloss.NewStyle().Width(34).Padding(1, 2).Render(brand)
	right := lipgloss.NewStyle().Width(72).Padding(1, 2).Render(strings.Join(rows, "\n"))
	card := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	if color {
		style = style.BorderForeground(border)
	}
	hints := "ESC Reset     F2 Command     F3 Session     ENTER Authenticate"
	rendered := style.Render(card)
	return rendered + "\n" + lipgloss.NewStyle().Width(lipgloss.Width(rendered)).Align(lipgloss.Center).Render(hints)
}
