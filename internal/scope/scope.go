package scope

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type Config struct {
	Includes []string
	Excludes []string
}

func (c Config) Validate() error {
	for _, list := range [][]string{c.Includes, c.Excludes} {
		seen := map[string]bool{}
		for _, v := range list {
			if _, err := normalize(v); err != nil {
				return err
			}
			if seen[v] {
				return fmt.Errorf("duplicate scope target %q", v)
			}
			seen[v] = true
		}
	}
	return nil
}

type Result string

const (
	Allow   Result = "ALLOW"
	Deny    Result = "DENY"
	Unknown Result = "UNKNOWN"
)

func Load(workbookRoot string) (Config, error) {
	b, err := os.ReadFile(filepath.Join(workbookRoot, "scope.yaml"))
	if err != nil {
		return Config{}, err
	}
	var c Config
	section := ""
	version := ""
	for _, line := range strings.Split(string(b), "\n") {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "version:") {
			version = strings.TrimSpace(strings.TrimPrefix(s, "version:"))
			continue
		}
		if s == "include:" {
			section = "include"
			continue
		}
		if s == "exclude:" {
			section = "exclude"
			continue
		}
		if strings.HasPrefix(s, "- ") {
			v := parseValue(strings.TrimSpace(strings.TrimPrefix(s, "- ")))
			if v == "[]" || v == "" {
				continue
			}
			if section == "include" {
				c.Includes = append(c.Includes, v)
			}
			if section == "exclude" {
				c.Excludes = append(c.Excludes, v)
			}
		}
	}
	if version != "1" {
		return Config{}, fmt.Errorf("unsupported scope schema version %q", version)
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return normalizeConfig(c)
}
func Save(workbookRoot string, c Config) error {
	normalized, err := normalizeConfig(c)
	if err != nil {
		return err
	}
	c = normalized
	var b strings.Builder
	b.WriteString("version: 1\ninclude:\n")
	for _, v := range c.Includes {
		fmt.Fprintf(&b, "  - %s\n", quote(v))
	}
	if len(c.Includes) == 0 {
		b.WriteString("  []\n")
	}
	b.WriteString("exclude:\n")
	for _, v := range c.Excludes {
		fmt.Fprintf(&b, "  - %s\n", quote(v))
	}
	if len(c.Excludes) == 0 {
		b.WriteString("  []\n")
	}
	path := filepath.Join(workbookRoot, "scope.yaml")
	tmp, err := os.CreateTemp(workbookRoot, ".scope-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.WriteString(b.String())
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err == nil {
		err = syncDir(workbookRoot)
	}
	return err
}
func Add(workbookRoot, target string, exclude bool) error {
	normalized, err := normalize(target)
	if err != nil {
		return err
	}
	c, err := Load(workbookRoot)
	if err != nil {
		return err
	}
	list := &c.Includes
	if exclude {
		list = &c.Excludes
	}
	for _, v := range *list {
		if v == normalized {
			return nil
		}
	}
	*list = append(*list, normalized)
	return Save(workbookRoot, c)
}

func Remove(workbookRoot, target string, exclude bool) error {
	normalized, err := normalize(target)
	if err != nil {
		return err
	}
	c, err := Load(workbookRoot)
	if err != nil {
		return err
	}
	list := c.Includes
	if exclude {
		list = c.Excludes
	}
	found := false
	out := list[:0]
	for _, v := range list {
		if v == normalized {
			found = true
			continue
		}
		out = append(out, v)
	}
	if !found {
		return fmt.Errorf("scope target %q not found", normalized)
	}
	if exclude {
		c.Excludes = out
	} else {
		c.Includes = out
	}
	return Save(workbookRoot, c)
}
func Evaluate(c Config, target string) Result {
	target = strings.TrimSpace(target)
	if target == "" {
		return Unknown
	}
	for _, x := range c.Excludes {
		if matches(x, target) {
			return Deny
		}
	}
	for _, x := range c.Includes {
		if matches(x, target) {
			return Allow
		}
	}
	return Unknown
}
func matches(rule, target string) bool {
	if rule == target {
		return true
	}
	rIP := net.ParseIP(rule)
	tIP := net.ParseIP(target)
	if rIP != nil && tIP != nil {
		return rIP.Equal(tIP)
	}
	if _, n, err := net.ParseCIDR(rule); err == nil && tIP != nil {
		return n.Contains(tIP)
	}
	host := target
	if u, err := url.Parse(target); err == nil && u.Hostname() != "" {
		host = u.Hostname()
	}
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	rule = strings.TrimSuffix(strings.ToLower(rule), ".")
	if u, err := url.Parse(rule); err == nil && u.Hostname() != "" {
		rule = strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	}
	if strings.HasPrefix(rule, "*.") {
		return strings.HasSuffix(host, rule[1:]) && host != rule[2:]
	}
	return host == rule
}

func normalizeConfig(c Config) (Config, error) {
	out := Config{Includes: make([]string, 0, len(c.Includes)), Excludes: make([]string, 0, len(c.Excludes))}
	for _, v := range c.Includes {
		n, err := normalize(v)
		if err != nil {
			return Config{}, err
		}
		out.Includes = append(out.Includes, n)
	}
	for _, v := range c.Excludes {
		n, err := normalize(v)
		if err != nil {
			return Config{}, err
		}
		out.Excludes = append(out.Excludes, n)
	}
	if err := out.Validate(); err != nil {
		return Config{}, err
	}
	return out, nil
}

func normalize(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.IndexFunc(value, unicode.IsSpace) >= 0 || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("scope target must be non-empty and contain no whitespace")
	}
	if ip := net.ParseIP(value); ip != nil {
		return ip.String(), nil
	}
	if _, network, err := net.ParseCIDR(value); err == nil {
		return network.String(), nil
	}
	if strings.HasPrefix(value, "*.") {
		return "*." + strings.ToLower(strings.TrimSuffix(value[2:], ".")), nil
	}
	return strings.ToLower(strings.TrimSuffix(value, ".")), nil
}

func parseValue(value string) string {
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return strings.ReplaceAll(value[1:len(value)-1], "''", "'")
	}
	return value
}

func quote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func syncDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
