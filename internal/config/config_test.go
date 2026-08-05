package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathsUseXDG(t *testing.T) {
	values := map[string]string{"XDG_CONFIG_HOME": "/tmp/config", "XDG_STATE_HOME": "/tmp/state", "HOME": "/home/test"}
	get := func(key string) string { return values[key] }
	configFile, stateDir := Paths(get)
	if configFile != "/tmp/config/pensuse/config.yaml" || stateDir != "/tmp/state/pensuse" {
		t.Fatalf("unexpected paths: %q %q", configFile, stateDir)
	}
}

func TestPathsDefaultToHome(t *testing.T) {
	get := func(key string) string {
		if key == "HOME" {
			return "/home/test"
		}
		return ""
	}
	configFile, stateDir := Paths(get)
	if configFile != "/home/test/.config/pensuse/config.yaml" || stateDir != "/home/test/.local/state/pensuse" {
		t.Fatalf("unexpected defaults: %q %q", configFile, stateDir)
	}
}

func TestEffectiveConfigPrecedence(t *testing.T) {
	root := t.TempDir()
	configHome := filepath.Join(root, "config")
	stateHome := filepath.Join(root, "state")
	if err := os.MkdirAll(filepath.Join(configHome, "pensuse"), 0700); err != nil {
		t.Fatal(err)
	}
	configFile := filepath.Join(configHome, "pensuse", "config.yaml")
	if err := os.WriteFile(configFile, []byte("schema: pensuse.config.v1\nworkbook_root: /from-file\nprofile_dir: /profiles\n"), 0600); err != nil {
		t.Fatal(err)
	}
	values := map[string]string{"XDG_CONFIG_HOME": configHome, "XDG_STATE_HOME": stateHome, "HOME": root, "PENSUSE_WORKBOOK_ROOT": "/from-env"}
	settings, err := Effective(func(key string) string { return values[key] }, func(path string) error { _, err := os.Stat(path); return err })
	if err != nil {
		t.Fatal(err)
	}
	if settings.WorkbookRoot != "/from-env" || settings.ProfileDir != "/profiles" {
		t.Fatalf("precedence: %+v", settings)
	}
}
