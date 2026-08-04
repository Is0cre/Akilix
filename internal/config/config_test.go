package config

import "testing"

func TestPathsUseXDG(t *testing.T) {
	values := map[string]string{"XDG_CONFIG_HOME": "/tmp/config", "XDG_STATE_HOME": "/tmp/state", "HOME": "/home/test"}
	get := func(key string) string { return values[key] }
	configFile, stateDir := Paths(get)
	if configFile != "/tmp/config/pensuse/config.yaml" || stateDir != "/tmp/state/pensuse" {
		t.Fatalf("unexpected paths: %q %q", configFile, stateDir)
	}
}

func TestPathsDefaultToHome(t *testing.T) {
	get := func(key string) string { if key == "HOME" { return "/home/test" }; return "" }
	configFile, stateDir := Paths(get)
	if configFile != "/home/test/.config/pensuse/config.yaml" || stateDir != "/home/test/.local/state/pensuse" {
		t.Fatalf("unexpected defaults: %q %q", configFile, stateDir)
	}
}
