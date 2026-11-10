// Package snippets validates multi-language code snippet examples under
// content/docs/code/ for region tag formatting, leaked author notes, balanced
// quotes, and language syntax.
package snippets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/report"
)

const domainName = "snippets"

var (
	regionStartRe = regexp.MustCompile(`\[START\s+([^\]]+)\]`)
	regionEndRe   = regexp.MustCompile(`\[END\s+([^\]]+)\]`)
	regionTagIDRe = regexp.MustCompile(`^[a-z][a-z0-9_]*[a-z0-9]$`)
)

// Check inspects all code snippet files under repoRoot/content/docs/code for
// style guide compliance and syntax errors.
func Check(repoRoot string) ([]report.Finding, error) {
	codeDir := filepath.Join(repoRoot, "content", "docs", "code")

	var findings []report.Finding
	for _, lang := range report.Languages() {
		langDir := filepath.Join(codeDir, lang.Folder)
		entries, err := os.ReadDir(langDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), "."+lang.Ext) {
				continue
			}
			codeFile := filepath.Join(langDir, entry.Name())
			rel, relErr := filepath.Rel(repoRoot, codeFile)
			if relErr != nil {
				rel = codeFile
			}
			rel = filepath.ToSlash(rel)

			raw, readErr := os.ReadFile(codeFile)
			if readErr != nil {
				return nil, fmt.Errorf("reading %q: %w", rel, readErr)
			}
			content := string(raw)

			findings = append(findings, checkAuthorNotes(rel, content)...)
			findings = append(findings, checkRegionTags(rel, content)...)
			findings = append(findings, checkBalancedQuotes(rel, lang.Folder, content)...)
			findings = append(findings, checkLanguageSyntax(rel, lang.Folder, codeFile)...)
		}
	}

	return findings, nil
}

func checkAuthorNotes(relFile, content string) []report.Finding {
	if !strings.Contains(content, "NOTE FOR SAMPLE AUTHOR") && !strings.Contains(content, "NOTE TO SAMPLE AUTHOR") {
		return nil
	}
	return []report.Finding{
		{
			Domain:   domainName,
			Rule:     "leaked-author-note",
			Severity: report.SeverityError,
			File:     relFile,
			Message:  `contains leaked internal "NOTE FOR/TO SAMPLE AUTHOR" comment`,
		},
	}
}

type unclosedTag struct {
	tag  string
	line int
}

func checkRegionTags(relFile, content string) []report.Finding {
	var findings []report.Finding
	starts := make(map[string]int)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		lineNum := i + 1
		for _, m := range regionStartRe.FindAllStringSubmatch(line, -1) {
			tag := strings.TrimSpace(m[1])
			if !regionTagIDRe.MatchString(tag) {
				findings = append(findings, report.Finding{
					Domain:   domainName,
					Rule:     "region-tag-format",
					Severity: report.SeverityError,
					File:     relFile,
					Line:     lineNum,
					Message:  fmt.Sprintf("region tag %q must be snake_case matching [a-z][a-z0-9_]*[a-z0-9]", tag),
				})
			}
			starts[tag] = lineNum
		}
		for _, m := range regionEndRe.FindAllStringSubmatch(line, -1) {
			tag := strings.TrimSpace(m[1])
			if _, ok := starts[tag]; !ok {
				findings = append(findings, report.Finding{
					Domain:   domainName,
					Rule:     "region-tag-mismatch",
					Severity: report.SeverityError,
					File:     relFile,
					Line:     lineNum,
					Message:  fmt.Sprintf("[END %s] has no matching [START %s]", tag, tag),
				})
				continue
			}
			delete(starts, tag)
		}
	}

	var remaining []unclosedTag
	for tag, lineNum := range starts {
		remaining = append(remaining, unclosedTag{tag: tag, line: lineNum})
	}
	sort.Slice(remaining, func(i, j int) bool {
		if remaining[i].line != remaining[j].line {
			return remaining[i].line < remaining[j].line
		}
		return remaining[i].tag < remaining[j].tag
	})

	for _, u := range remaining {
		findings = append(findings, report.Finding{
			Domain:   domainName,
			Rule:     "region-tag-mismatch",
			Severity: report.SeverityError,
			File:     relFile,
			Line:     u.line,
			Message:  fmt.Sprintf("[START %s] has no matching [END %s]", u.tag, u.tag),
		})
	}
	return findings
}

// checkBalancedQuotes scans C#, Java, Go, and Terraform files for single-line
// double-quoted string literals that reach an unescaped newline or EOF without
// closing, while ignoring line/block comments, character literals, backtick raw
// strings, and triple-quoted multiline text blocks.
func checkBalancedQuotes(relFile, folder, content string) []report.Finding {
	switch folder {
	case "csharp", "java", "go", "terraform":
	default:
		return nil
	}

	lines := strings.Split(content, "\n")
	var findings []report.Finding

	type scanState int
	const (
		stateNormal scanState = iota
		stateLineComment
		stateBlockComment
		stateSingleQuote
		stateBacktick
		stateTripleQuote
		stateDoubleQuote
		stateVerbatimDoubleQuote
	)

	state := stateNormal
	openLine := 1
	lineNum := 1

	for i := 0; i < len(content); i++ {
		ch := content[i]
		if ch == '\n' {
			if state == stateDoubleQuote {
				lineText := strings.TrimSpace(lines[openLine-1])
				findings = append(findings, report.Finding{
					Domain:   domainName,
					Rule:     "unbalanced-quotes",
					Severity: report.SeverityError,
					File:     relFile,
					Line:     openLine,
					Message:  fmt.Sprintf("unbalanced double quote in line %q", lineText),
				})
				state = stateNormal
			} else if state == stateLineComment || state == stateSingleQuote {
				state = stateNormal
			}
			lineNum++
			continue
		}

		switch state {
		case stateLineComment:
			continue

		case stateBlockComment:
			if ch == '*' && i+1 < len(content) && content[i+1] == '/' {
				state = stateNormal
				i++
			}

		case stateSingleQuote:
			if ch == '\\' && i+1 < len(content) {
				i++
				continue
			}
			if ch == '\'' {
				state = stateNormal
			}

		case stateBacktick:
			if ch == '`' {
				state = stateNormal
			}

		case stateTripleQuote:
			if ch == '"' && i+2 < len(content) && content[i+1] == '"' && content[i+2] == '"' {
				state = stateNormal
				i += 2
			}

		case stateVerbatimDoubleQuote:
			if ch == '"' {
				if i+1 < len(content) && content[i+1] == '"' {
					i++
					continue
				}
				state = stateNormal
			}

		case stateDoubleQuote:
			if ch == '\\' && i+1 < len(content) {
				i++
				continue
			}
			if ch == '"' {
				state = stateNormal
			}

		case stateNormal:
			if ch == '/' && i+1 < len(content) {
				if content[i+1] == '/' {
					state = stateLineComment
					i++
					continue
				}
				if content[i+1] == '*' {
					state = stateBlockComment
					i++
					continue
				}
			}
			if ch == '#' && folder == "terraform" {
				state = stateLineComment
				continue
			}
			if ch == '\'' {
				state = stateSingleQuote
				continue
			}
			if ch == '`' && folder == "go" {
				state = stateBacktick
				continue
			}
			if ch == '@' && folder == "csharp" {
				if i+1 < len(content) && content[i+1] == '"' {
					state = stateVerbatimDoubleQuote
					openLine = lineNum
					i++
					continue
				}
				if i+2 < len(content) && content[i+1] == '$' && content[i+2] == '"' {
					state = stateVerbatimDoubleQuote
					openLine = lineNum
					i += 2
					continue
				}
			}
			if ch == '$' && folder == "csharp" && i+2 < len(content) && content[i+1] == '@' && content[i+2] == '"' {
				state = stateVerbatimDoubleQuote
				openLine = lineNum
				i += 2
				continue
			}
			if ch == '"' {
				if i+2 < len(content) && content[i+1] == '"' && content[i+2] == '"' {
					state = stateTripleQuote
					openLine = lineNum
					i += 2
					continue
				}
				state = stateDoubleQuote
				openLine = lineNum
			}
		}
	}

	if state == stateDoubleQuote {
		lineText := strings.TrimSpace(lines[openLine-1])
		findings = append(findings, report.Finding{
			Domain:   domainName,
			Rule:     "unbalanced-quotes",
			Severity: report.SeverityError,
			File:     relFile,
			Line:     openLine,
			Message:  fmt.Sprintf("unbalanced double quote in line %q", lineText),
		})
	}

	return findings
}

func checkLanguageSyntax(relFile, folder, codeFile string) []report.Finding {
	if folder != "ruby" {
		return nil
	}
	if _, err := exec.LookPath("ruby"); err != nil {
		return nil
	}
	cmd := exec.Command("ruby", "-c", codeFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		return []report.Finding{
			{
				Domain:   domainName,
				Rule:     "syntax-error",
				Severity: report.SeverityError,
				File:     relFile,
				Message:  fmt.Sprintf("ruby syntax error: %s", strings.TrimSpace(string(out))),
			},
		}
	}
	return nil
}
