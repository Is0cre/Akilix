package statusbar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Is0cre/Akilix/internal/activity"
	"github.com/Is0cre/Akilix/internal/journal"
	"github.com/Is0cre/Akilix/internal/scope"
	"github.com/Is0cre/Akilix/internal/workbook"
)

type State struct {
	Root     string `json:"root"`
	Workbook string `json:"workbook"`
}
type Metrics struct {
	Workbook string
	Scope    string
	Running  int
	Failed   int
	Hosts    int
	Ports    int
	Dropped  int
}
type Block struct {
	FullText   string `json:"full_text"`
	Color      string `json:"color,omitempty"`
	Background string `json:"background,omitempty"`
	Separator  bool   `json:"separator"`
}

func Activate(runtimeDir, root, name string) error {
	if runtimeDir == "" {
		return fmt.Errorf("XDG_RUNTIME_DIR is unavailable")
	}
	if _, err := workbook.Open(root, name); err != nil {
		return err
	}
	dir := filepath.Join(runtimeDir, "akilix")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, _ := json.Marshal(State{Root: root, Workbook: name})
	tmp, err := os.CreateTemp(dir, ".active-")
	if err != nil {
		return err
	}
	path := tmp.Name()
	defer os.Remove(path)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(append(b, '\n'))
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(path, filepath.Join(dir, "active-workbook.json"))
	}
	return err
}

// Active returns the explicitly selected, per-user workbook context. It does
// not activate logging, scope enforcement, or any network operation.
func Active(runtimeDir string) (State, error) {
	if runtimeDir == "" {
		return State{}, fmt.Errorf("XDG_RUNTIME_DIR is unavailable")
	}
	path := filepath.Join(runtimeDir, "akilix", "active-workbook.json")
	info, err := os.Lstat(path)
	if err != nil {
		return State{}, err
	}
	if !info.Mode().IsRegular() {
		return State{}, fmt.Errorf("invalid active workbook state %q", path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(b, &state); err != nil {
		return State{}, err
	}
	if _, err := workbook.Open(state.Root, state.Workbook); err != nil {
		return State{}, err
	}
	return state, nil
}

// Deactivate removes only the ephemeral UI context. Workbook files and
// managed records are never changed.
func Deactivate(runtimeDir string) error {
	if runtimeDir == "" {
		return fmt.Errorf("XDG_RUNTIME_DIR is unavailable")
	}
	path := filepath.Join(runtimeDir, "akilix", "active-workbook.json")
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("invalid active workbook state %q", path)
	}
	return os.Remove(path)
}

func Current(runtimeDir string) (Metrics, error) {
	state, err := Active(runtimeDir)
	if err != nil {
		return Metrics{}, err
	}
	config, err := scope.Load(filepath.Join(state.Root, state.Workbook))
	if err != nil {
		return Metrics{}, err
	}
	events, err := activity.List(filepath.Join(state.Root, state.Workbook))
	if err != nil {
		return Metrics{}, err
	}
	discoveries, err := journal.Summarize(filepath.Join(state.Root, state.Workbook))
	if err != nil {
		return Metrics{}, err
	}
	running := map[string]bool{}
	failed := 0
	for _, event := range events {
		switch event.Phase {
		case "STARTED":
			running[event.InvocationID] = true
		case "COMPLETED":
			delete(running, event.InvocationID)
		case "FAILED":
			delete(running, event.InvocationID)
			failed++
		}
	}
	scopeText := "UNDECLARED"
	if len(config.Includes) > 0 {
		values := append([]string(nil), config.Includes...)
		sort.Strings(values)
		scopeText = strings.Join(values, ",")
	}
	return Metrics{
		Workbook: state.Workbook,
		Scope:    scopeText,
		Running:  len(running),
		Failed:   failed,
		Hosts:    discoveries.DiscoveredHosts,
		Ports:    discoveries.DiscoveredPorts,
		Dropped:  discoveries.DroppedResults,
	}, nil
}

func Blocks(m Metrics) []Block {
	warn := "#8ead55"
	if m.Failed > 0 {
		warn = "#e06c75"
	}
	scopeBlock := Block{FullText: " 🚨 SCOPE UNDECLARED ", Color: "#ff5f00", Separator: false}
	if m.Scope != "UNDECLARED" {
		scopeBlock = Block{FullText: " 🎯 SCOPE ACTIVE · " + m.Scope + " ", Color: "#87ff00", Separator: false}
	}
	return []Block{
		{FullText: " 📂 " + m.Workbook + " ", Color: "#e8e6dd", Background: "#657a3e", Separator: false},
		scopeBlock,
		{FullText: fmt.Sprintf(" 📡 HOSTS %d · PORTS %d · DROPPED %d ", m.Hosts, m.Ports, m.Dropped), Color: "#72b7c9", Separator: false},
		{FullText: fmt.Sprintf(" ⚠ FAILED %d ", m.Failed), Color: warn, Separator: false},
		{FullText: fmt.Sprintf(" 📦 RUNNING %d ", m.Running), Color: "#c68a2b", Separator: false},
	}
}

// TimeBlocks renders local calendar and clock state. The caller supplies time
// so formatting is deterministic in tests and follows the configured host zone.
func TimeBlocks(now time.Time) []Block {
	return []Block{
		{FullText: " 📅 " + now.Format("Mon 02 Jan") + " ", Color: "#e8e6dd", Separator: false},
		{FullText: " 🕒 " + now.Format("15:04 MST") + " ", Color: "#0b1114", Background: "#8ead55", Separator: false},
	}
}
