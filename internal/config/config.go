package config

import (
	"os"
	"path/filepath"
)

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
	return filepath.Join(configHome, "pensuse", "config.yaml"), filepath.Join(stateHome, "pensuse")
}

func UserPaths() (string, string) { return Paths(os.Getenv) }
