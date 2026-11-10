// Package report defines diagnostic findings, language metadata, and output
// formatters shared across catalog validation domains.
package report

import (
	"encoding/json"
	"fmt"
	"io"
)

// Severity indicates whether a validation finding is a blocking error or an
// advisory warning.
type Severity string

const (
	// SeverityError represents a validation failure that causes a non-zero exit.
	SeverityError Severity = "error"
	// SeverityWarning represents a non-blocking advisory finding.
	SeverityWarning Severity = "warning"
)

// Finding represents a single validation issue discovered in the repository.
type Finding struct {
	Domain   string   `json:"domain"`
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity"`
	File     string   `json:"file"`
	Line     int      `json:"line,omitempty"`
	Message  string   `json:"message"`
}

// Language describes a programming language tab supported by the
// guideline_examples shortcode.
type Language struct {
	Name   string
	Folder string
	Ext    string
}

// Languages returns the canonical list of languages supported by
// layouts/_shortcodes/guideline_examples.html.
func Languages() []Language {
	return []Language{
		{Name: "Python", Folder: "python", Ext: "py"},
		{Name: "Node.js", Folder: "node", Ext: "js"},
		{Name: "Java", Folder: "java", Ext: "java"},
		{Name: "Go", Folder: "go", Ext: "go"},
		{Name: "C#", Folder: "csharp", Ext: "cs"},
		{Name: "Ruby", Folder: "ruby", Ext: "rb"},
		{Name: "PHP", Folder: "php", Ext: "php"},
		{Name: "Terraform", Folder: "terraform", Ext: "tf"},
	}
}

// HasErrors reports whether any finding in the slice has SeverityError.
func HasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.Severity == SeverityError {
			return true
		}
	}
	return false
}

// WriteText formats findings as human-readable grouped output to w.
func WriteText(w io.Writer, findings []Finding) error {
	var warnings []Finding
	var errs []Finding
	for _, f := range findings {
		if f.Severity == SeverityWarning {
			warnings = append(warnings, f)
			continue
		}
		errs = append(errs, f)
	}

	if len(warnings) > 0 {
		if _, err := fmt.Fprintf(w, "Warnings (%d):\n", len(warnings)); err != nil {
			return fmt.Errorf("writing warning header: %w", err)
		}
		for _, f := range warnings {
			if _, err := fmt.Fprintf(w, "  [WARN] [%s/%s] %s\n", f.Domain, f.Rule, formatLocation(f)); err != nil {
				return fmt.Errorf("writing warning finding: %w", err)
			}
		}
	}

	if len(errs) > 0 {
		if _, err := fmt.Fprintf(w, "Errors (%d):\n", len(errs)); err != nil {
			return fmt.Errorf("writing error header: %w", err)
		}
		for _, f := range errs {
			if _, err := fmt.Fprintf(w, "  [ERROR] [%s/%s] %s\n", f.Domain, f.Rule, formatLocation(f)); err != nil {
				return fmt.Errorf("writing error finding: %w", err)
			}
		}
		return nil
	}

	if _, err := fmt.Fprintln(w, "All catalog checks passed with 0 errors!"); err != nil {
		return fmt.Errorf("writing success summary: %w", err)
	}
	return nil
}

// WriteJSON serializes findings as indented JSON to w.
func WriteJSON(w io.Writer, findings []Finding) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if findings == nil {
		findings = []Finding{}
	}
	if err := enc.Encode(findings); err != nil {
		return fmt.Errorf("encoding findings as json: %w", err)
	}
	return nil
}

func formatLocation(f Finding) string {
	if f.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", f.File, f.Line, f.Message)
	}
	if f.File != "" {
		return fmt.Sprintf("%s: %s", f.File, f.Message)
	}
	return f.Message
}
