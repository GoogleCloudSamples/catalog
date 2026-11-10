// Package shortcodes validates Hugo shortcode calls in Markdown files against
// the multi-language code snippet files under content/docs/code/.
package shortcodes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/report"
)

const domainName = "shortcodes"

var guidelineRe = regexp.MustCompile(`\{\{[<%]\s*guideline_examples\s+"([^"]+)"\s*[>%]\}\}`)
var (
	tabRe = regexp.MustCompile(`\{\{[<%]\s*tab\s+name="([^"]+)"`)
)

// Check inspects Hugo content under repoRoot for guideline_examples shortcode
// calls and verifies exact-case file mappings, unreferenced code files, and
// invalid "# N/A" stubs.
func Check(repoRoot string) ([]report.Finding, error) {
	contentDir := filepath.Join(repoRoot, "content")
	codeDir := filepath.Join(contentDir, "docs", "code")

	referenced, err := collectReferencedGuidelines(contentDir)
	if err != nil {
		return nil, fmt.Errorf("scanning markdown shortcodes: %w", err)
	}

	var slugs []string
	for slug := range referenced {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	var findings []report.Finding
	langs := report.Languages()

	// 1. Check case-sensitive filename resolution for every referenced slug.
	for _, slug := range slugs {
		for _, lang := range langs {
			langDir := filepath.Join(codeDir, lang.Folder)
			for _, ext := range []string{lang.Ext, "md"} {
				expectedName := fmt.Sprintf("%s.%s", slug, ext)
				expectedPath := filepath.Join(langDir, expectedName)
				if fileExists(expectedPath) {
					continue
				}
				entries, readErr := os.ReadDir(langDir)
				if readErr != nil {
					continue
				}
				for _, entry := range entries {
					if !entry.IsDir() && strings.EqualFold(entry.Name(), expectedName) && entry.Name() != expectedName {
						relFile := filepath.ToSlash(filepath.Join("content", "docs", "code", lang.Folder, entry.Name()))
						findings = append(findings, report.Finding{
							Domain:   domainName,
							Rule:     "case-mismatch",
							Severity: report.SeverityError,
							File:     relFile,
							Message:  fmt.Sprintf("case mismatch for %s guideline %q: found %q, expected %q", lang.Name, slug, entry.Name(), expectedName),
						})
					}
				}
			}
		}
	}

	// 2. Check all code and companion note files for unreferenced guidelines and "# N/A" stubs.
	for _, lang := range langs {
		langDir := filepath.Join(codeDir, lang.Folder)
		entries, readErr := os.ReadDir(langDir)
		if readErr != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := filepath.Ext(entry.Name())
			isCodeFile := strings.EqualFold(ext, "."+lang.Ext)
			isNoteFile := strings.EqualFold(ext, ".md") &&
				!strings.EqualFold(entry.Name(), "_index.md") &&
				!strings.EqualFold(entry.Name(), "readme.md")
			if !isCodeFile && !isNoteFile {
				continue
			}
			codeFile := filepath.Join(langDir, entry.Name())
			rel, relErr := filepath.Rel(repoRoot, codeFile)
			if relErr != nil {
				rel = codeFile
			}
			rel = filepath.ToSlash(rel)

			slug := strings.ToLower(strings.TrimSuffix(entry.Name(), ext))
			if !referenced[slug] {
				findings = append(findings, report.Finding{
					Domain:   domainName,
					Rule:     "unreferenced-guideline",
					Severity: report.SeverityWarning,
					File:     rel,
					Message:  fmt.Sprintf("guideline %q is not referenced by any guideline_examples shortcode", slug),
				})
			}

			raw, fileErr := os.ReadFile(codeFile)
			if fileErr != nil {
				return nil, fmt.Errorf("reading %q: %w", rel, fileErr)
			}
			stripped := strings.TrimSpace(string(raw))
			if stripped == "# N/A" || stripped == "// N/A" || stripped == "/* N/A */" || stripped == "" {
				findings = append(findings, report.Finding{
					Domain:   domainName,
					Rule:     "na-stub-file",
					Severity: report.SeverityError,
					File:     rel,
					Message:  fmt.Sprintf("contains stub %q; omit the file so Hextra renders the 'Not applicable' callout", stripped),
				})
			}
		}
	}

	// 3. Check hand-written tab names, which single-language outputs filter by name.
	tabFindings, err := checkTabNames(repoRoot, contentDir, codeDir)
	if err != nil {
		return nil, err
	}
	findings = append(findings, tabFindings...)

	return findings, nil
}

// checkTabNames warns about {{< tab name="..." >}} names that are not a
// language name. Single-language outputs keep such tabs in every file, so a
// misspelled language ("NodeJS") leaks into the other languages' files.
func checkTabNames(repoRoot, contentDir, codeDir string) ([]report.Finding, error) {
	known := map[string]bool{}
	for _, lang := range report.Languages() {
		known[lang.Name] = true
	}
	var findings []report.Finding
	err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path == codeDir {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %q: %w", path, err)
		}
		rel, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			rel = path
		}
		for i, line := range strings.Split(string(raw), "\n") {
			for _, m := range tabRe.FindAllStringSubmatch(line, -1) {
				if known[m[1]] {
					continue
				}
				findings = append(findings, report.Finding{
					Domain:   domainName,
					Rule:     "unknown-tab-name",
					Severity: report.SeverityWarning,
					File:     filepath.ToSlash(rel),
					Line:     i + 1,
					Message:  fmt.Sprintf("tab %q is not a language name; single-language llms outputs keep it in every language's file", m[1]),
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning tab shortcodes: %w", err)
	}
	return findings, nil
}

func collectReferencedGuidelines(contentDir string) (map[string]bool, error) {
	referenced := make(map[string]bool)
	err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %q: %w", path, err)
		}
		for _, m := range guidelineRe.FindAllStringSubmatch(string(raw), -1) {
			referenced[m[1]] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return referenced, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
