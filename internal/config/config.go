package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const Schema = "akilix.config.v1"

type Settings struct {
	Schema       string `json:"schema"`
	ConfigFile   string `json:"config_file"`
	StateDir     string `json:"state_dir"`
	WorkbookRoot string `json:"workbook_root"`
	ProfileDir   string `json:"profile_dir"`
}

func (s Settings) Validate() error {
	if s.Schema != Schema || s.ConfigFile == "" || s.StateDir == "" || s.WorkbookRoot == "" || s.ProfileDir == "" {
		return fmt.Errorf("invalid Akilix configuration")
	}
	return nil
}

// Paths returns the user configuration and state locations using XDG rules.
// Empty environment values are ignored so compiled defaults remain usable.
func Paths(environ func(string) string) (configFile, stateDir string) {
	configHome := environ("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(environ("HOME"), ".config")
	}
	stateHome := environ("XDG_STATE_HOME")
	if stateHome == "" {
		stateHome = filepath.Join(environ("HOME"), ".local", "state")
	}
	return filepath.Join(configHome, "akilix", "config.yaml"), filepath.Join(stateHome, "akilix")
}

func UserPaths() (string, string) { return Paths(os.Getenv) }

func Effective(environ func(string) string, stat func(string) error) (Settings, error) {
	configFile, stateDir := Paths(environ)
	settings := Settings{Schema: Schema, ConfigFile: configFile, StateDir: stateDir, WorkbookRoot: filepath.Join(stateDir, "workbooks"), ProfileDir: "/usr/share/akilix/profiles"}
	if err := stat(configFile); err == nil {
		loaded, err := load(configFile)
		if err != nil {
			return Settings{}, err
		}
		if loaded.WorkbookRoot != "" {
			settings.WorkbookRoot = loaded.WorkbookRoot
		}
		if loaded.ProfileDir != "" {
			settings.ProfileDir = loaded.ProfileDir
		}
	} else if !os.IsNotExist(err) {
		return Settings{}, err
	}
	if value := environ("AKILIX_WORKBOOK_ROOT"); value != "" {
		settings.WorkbookRoot = value
	}
	if value := environ("AKILIX_PROFILE_DIR"); value != "" {
		settings.ProfileDir = value
	}
	return settings, settings.Validate()
}

func Load(path string) (Settings, error) { return load(path) }

func load(path string) (Settings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}
	var s Settings
	for _, line := range strings.Split(string(b), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch key {
		case "schema":
			s.Schema = value
		case "workbook_root":
			s.WorkbookRoot = value
		case "profile_dir":
			s.ProfileDir = value
		}
	}
	if s.Schema == "" {
		s.Schema = Schema
	}
	return s, nil
}
