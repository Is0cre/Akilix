package container

import (
	"fmt"
	"os"
	"path/filepath"
)

// OriginalEvidenceMount exposes only the canonical original-evidence tree at
// a fixed container path. It rejects replacement by a symlink, including a
// symlink in an intermediate workbook path.
func OriginalEvidenceMount(workbookRoot string) (Mount, error) {
	absRoot, err := filepath.Abs(workbookRoot)
	if err != nil {
		return Mount{}, err
	}
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return Mount{}, fmt.Errorf("resolve workbook root: %w", err)
	}
	source := filepath.Join(absRoot, "evidence", "original")
	info, err := os.Lstat(source)
	if err != nil {
		return Mount{}, fmt.Errorf("inspect original evidence directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Mount{}, fmt.Errorf("original evidence path is not a real directory")
	}
	realSource, err := filepath.EvalSymlinks(source)
	if err != nil {
		return Mount{}, fmt.Errorf("resolve original evidence directory: %w", err)
	}
	want := filepath.Join(realRoot, "evidence", "original")
	if realSource != want {
		return Mount{}, fmt.Errorf("original evidence path crosses a symlink boundary")
	}
	return Mount{Source: source, Destination: "/workbook/evidence/original", ReadOnly: true, OriginalEvidence: true}, nil
}
