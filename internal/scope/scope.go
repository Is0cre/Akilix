package scope

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Includes []string
	Excludes []string
}

func (c Config) Validate() error {
	seen := map[string]bool{}
	for _, list := range [][]string{c.Includes, c.Excludes} {
		for _, v := range list {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("scope target cannot be empty")
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
	for _, line := range strings.Split(string(b), "\n") {
		s := strings.TrimSpace(line)
		if s == "include:" {
			section = "include"
			continue
		}
		if s == "exclude:" {
			section = "exclude"
			continue
		}
		if strings.HasPrefix(s, "- ") {
			v := strings.TrimSpace(strings.TrimPrefix(s, "- "))
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
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}
func Save(workbookRoot string, c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("version: 1\ninclude:\n")
	for _, v := range c.Includes {
		fmt.Fprintf(&b, "  - %s\n", v)
	}
	if len(c.Includes) == 0 {
		b.WriteString("  []\n")
	}
	b.WriteString("exclude:\n")
	for _, v := range c.Excludes {
		fmt.Fprintf(&b, "  - %s\n", v)
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
	return err
}
func Add(workbookRoot, target string, exclude bool) error {
	c, err := Load(workbookRoot)
	if err != nil {
		return err
	}
	list := &c.Includes
	if exclude {
		list = &c.Excludes
	}
	for _, v := range *list {
		if v == target {
			return nil
		}
	}
	*list = append(*list, target)
	return Save(workbookRoot, c)
}

func Remove(workbookRoot, target string, exclude bool) error {
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
		if v == target {
			found = true
			continue
		}
		out = append(out, v)
	}
	if !found {
		return fmt.Errorf("scope target %q not found", target)
	}
	if exclude {
		c.Excludes = out
	} else {
		c.Includes = out
	}
	return Save(workbookRoot, c)
}
func Evaluate(c Config, target string) Result {
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
	if strings.HasPrefix(rule, "*.") {
		return strings.HasSuffix(host, rule[1:]) && host != rule[2:]
	}
	return host == rule
}
