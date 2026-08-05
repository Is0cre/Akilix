package completion

import "testing"

func TestScriptsContainCoreCommands(t *testing.T) {
	for name, script := range map[string]string{"zsh": Zsh, "bash": Bash} {
		for _, command := range []string{"workbook", "scope", "evidence", "completion"} {
			if len(script) == 0 || !contains(script, command) {
				t.Errorf("%s completion missing %s", name, command)
			}
		}
	}
}
func contains(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
