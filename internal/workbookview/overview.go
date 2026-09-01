package workbookview

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Is0cre/Akilix/internal/acquire"
	"github.com/Is0cre/Akilix/internal/evidence"
	"github.com/Is0cre/Akilix/internal/invocation"
	"github.com/Is0cre/Akilix/internal/logpolicy"
	"github.com/Is0cre/Akilix/internal/scope"
	"github.com/Is0cre/Akilix/internal/workbook"
)

type Overview struct {
	Name              string           `json:"name"`
	ID                string           `json:"id"`
	Status            string           `json:"status"`
	Created           string           `json:"created"`
	Root              string           `json:"root"`
	ScopeIncludes     int              `json:"scope_includes"`
	ScopeExcludes     int              `json:"scope_excludes"`
	Evidence          int              `json:"evidence"`
	Invocations       int              `json:"invocations"`
	FailedInvocations int              `json:"failed_invocations"`
	Acquisitions      int              `json:"acquisitions"`
	RecoveryRequired  int              `json:"recovery_required"`
	Logging           logpolicy.Policy `json:"logging"`
	Sections          []Section        `json:"sections"`
}

type Section struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

var sectionPaths = map[string]string{
	"artifacts":         "artifacts",
	"evidence":          "evidence",
	"findings":          "findings",
	"hardware":          "hardware",
	"logs":              "logs",
	"notes":             "notes",
	"original-evidence": "evidence/original",
	"reports":           "reports",
	"timeline":          "timeline",
	"tool-output":       "tool-output",
}

func Build(root, name string) (Overview, error) {
	if err := workbook.ValidateLayout(root, name); err != nil {
		return Overview{}, err
	}
	metadata, err := workbook.Open(root, name)
	if err != nil {
		return Overview{}, err
	}
	workbookRoot, err := filepath.Abs(filepath.Join(root, name))
	if err != nil {
		return Overview{}, err
	}
	scopeConfig, err := scope.Load(workbookRoot)
	if err != nil {
		return Overview{}, err
	}
	policy, err := logpolicy.Load(workbookRoot)
	if err != nil {
		return Overview{}, err
	}
	evidenceRecords, err := evidence.List(workbookRoot)
	if err != nil {
		return Overview{}, err
	}
	invocations, err := invocation.List(workbookRoot)
	if err != nil {
		return Overview{}, err
	}
	acquisitions, err := acquire.ImageStatus(workbookRoot, "")
	if err != nil {
		return Overview{}, err
	}
	recoveryRequired := 0
	for _, operation := range acquisitions {
		if operation.RecoveryRequired {
			recoveryRequired++
		}
	}
	failed := 0
	for _, record := range invocations {
		if record.Status == "failed" {
			failed++
		}
	}
	sections := make([]Section, 0, len(sectionPaths)+1)
	sections = append(sections, Section{Name: "root", Path: workbookRoot})
	names := make([]string, 0, len(sectionPaths))
	for section := range sectionPaths {
		names = append(names, section)
	}
	sort.Strings(names)
	for _, section := range names {
		sections = append(sections, Section{Name: section, Path: filepath.Join(workbookRoot, sectionPaths[section])})
	}
	return Overview{
		Name: metadata.Name, ID: metadata.ID, Status: metadata.Status,
		Created: metadata.Created.Format(time.RFC3339Nano), Root: workbookRoot,
		ScopeIncludes: len(scopeConfig.Includes), ScopeExcludes: len(scopeConfig.Excludes),
		Evidence: len(evidenceRecords), Invocations: len(invocations), FailedInvocations: failed,
		Acquisitions: len(acquisitions), RecoveryRequired: recoveryRequired,
		Logging: policy, Sections: sections,
	}, nil
}

func Path(root, name, section string) (string, error) {
	if _, err := workbook.Open(root, name); err != nil {
		return "", err
	}
	relative := ""
	if section != "" && section != "root" {
		var ok bool
		relative, ok = sectionPaths[section]
		if !ok {
			return "", fmt.Errorf("unknown workbook section %q", section)
		}
	}
	path, err := filepath.Abs(filepath.Join(root, name, relative))
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("invalid workbook section %q", section)
	}
	return path, nil
}
