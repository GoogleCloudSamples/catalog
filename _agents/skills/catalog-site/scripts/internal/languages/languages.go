// Package languages checks that every language in
// layouts/_partials/samples/languages.html is set up consistently: it
// matches report.Languages, and has its llms-full output format and
// templates.
package languages

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/report"
)

const domainName = "languages"

var (
	languageRe = regexp.MustCompile(`\(dict "name" "([^"]+)" "lang" "[^"]+" "ext" "([^"]+)" "folder" "([^"]+)"\)`)
	folderRe   = regexp.MustCompile(`"folder" "([^"]+)"`)
	// homeOutputsRe captures the output formats listed in outputs.home.
	homeOutputsRe = regexp.MustCompile(`(?m)^\s*home:\s*\[([^\]]*)\]`)
	// yamlBlockKeyRe captures indented YAML keys with no inline value, such as
	// output format names under outputFormats.
	yamlBlockKeyRe = regexp.MustCompile(`(?m)^\s+([A-Za-z0-9_-]+):\s*$`)
)

// Check validates the language list under repoRoot.
func Check(repoRoot string) ([]report.Finding, error) {
	findings := checkLanguageList(repoRoot)
	outputs, err := checkLanguageOutputs(repoRoot)
	if err != nil {
		return nil, err
	}
	return append(findings, outputs...), nil
}

// checkLanguageList reports drift between layouts/_partials/samples/languages.html
// and report.Languages, which must list the same languages in the same order.
func checkLanguageList(repoRoot string) []report.Finding {
	const rel = "layouts/_partials/samples/languages.html"
	raw, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(rel)))
	if err != nil {
		return nil
	}
	var got []string
	for _, m := range languageRe.FindAllStringSubmatch(string(raw), -1) {
		got = append(got, fmt.Sprintf("%s/%s/%s", m[1], m[3], m[2]))
	}
	var want []string
	for _, lang := range report.Languages() {
		want = append(want, fmt.Sprintf("%s/%s/%s", lang.Name, lang.Folder, lang.Ext))
	}
	if strings.Join(got, ",") == strings.Join(want, ",") {
		return nil
	}
	return []report.Finding{{
		Domain:   domainName,
		Rule:     "language-list-drift",
		Severity: report.SeverityError,
		File:     rel,
		Message:  fmt.Sprintf("languages %v do not match report.Languages() %v (name/folder/ext, in order)", got, want),
	}}
}

// checkLanguageOutputs requires every language in languages.html to have
// its output format, home output entry, and one-line templates.
func checkLanguageOutputs(repoRoot string) ([]report.Finding, error) {
	const langRel = "layouts/_partials/samples/languages.html"
	raw, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(langRel)))
	if err != nil {
		return nil, nil // Not configured.
	}
	cfgRaw, err := os.ReadFile(filepath.Join(repoRoot, "hugo.yaml"))
	if err != nil {
		return nil, fmt.Errorf("reading hugo.yaml: %w", err)
	}
	cfg := string(cfgRaw)
	var findings []report.Finding
	homeLine := homeOutputsRe.FindStringSubmatch(cfg)
	home := map[string]bool{}
	if homeLine != nil {
		for _, f := range strings.Split(homeLine[1], ",") {
			home[strings.TrimSpace(f)] = true
		}
	}
	defined := map[string]bool{}
	for _, m := range yamlBlockKeyRe.FindAllStringSubmatch(cfg, -1) {
		defined[m[1]] = true
	}
	formats := []string{"llmsfull"}
	for _, m := range folderRe.FindAllStringSubmatch(string(raw), -1) {
		formats = append(formats, "llms"+m[1])
	}
	for _, f := range formats {
		if !defined[f] {
			findings = add(findings, "incomplete-language-outputs", "hugo.yaml", 0, "missing outputFormats.%s", f)
		}
		if !home[f] {
			findings = add(findings, "incomplete-language-outputs", "hugo.yaml", 0, "outputs.home does not include %s", f)
		}
		files := []string{"layouts/" + f + ".txt"}
		if f != "llmsfull" {
			files = append(files,
				"layouts/_shortcodes/guideline_examples."+f+".txt",
				"layouts/_shortcodes/tab."+f+".txt")
		}
		for _, rel := range files {
			if !fileExists(filepath.Join(repoRoot, filepath.FromSlash(rel))) {
				findings = add(findings, "incomplete-language-outputs", rel, 0, "missing template for output format %s", f)
			}
		}
	}
	return findings, nil
}

// add appends an error finding in this domain.
func add(findings []report.Finding, rule, file string, line int, format string, args ...any) []report.Finding {
	return append(findings, report.Finding{
		Domain:   domainName,
		Rule:     rule,
		Severity: report.SeverityError,
		File:     file,
		Line:     line,
		Message:  fmt.Sprintf(format, args...),
	})
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
