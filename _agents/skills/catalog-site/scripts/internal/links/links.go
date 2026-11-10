// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package links validates Markdown links, heading anchor targets, and content
// hygiene across Hugo Markdown files under content/.
package links

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

const domainName = "links"

var (
	headingIDRe   = regexp.MustCompile(`\{#([a-zA-Z0-9_-]+)\}`)
	headingTextRe = regexp.MustCompile(`(?m)^#{1,6}\s+(.+?)(?:\s*\{#[a-zA-Z0-9_-]+\})?\s*$`)
	mdLinkRe      = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	shortcodeRe   = regexp.MustCompile(`\{\{.*?\}\}`)
	formattingRe  = regexp.MustCompile(`[*_` + "`" + `]`)
	nonSlugRe     = regexp.MustCompile(`[^a-z0-9\s-]`)
	dashSpaceRe   = regexp.MustCompile(`[\s-]+`)
)

// Check inspects all Markdown files under repoRoot/content for broken relative
// links, missing heading anchors, and stray placeholder artifacts.
func Check(repoRoot string) ([]report.Finding, error) {
	contentDir := filepath.Join(repoRoot, "content")

	mdFiles, err := collectMarkdownFiles(contentDir)
	if err != nil {
		return nil, fmt.Errorf("walking content dir: %w", err)
	}

	anchorsByFile := make(map[string]map[string]bool, len(mdFiles))
	contentsByFile := make(map[string]string, len(mdFiles))

	var findings []report.Finding

	for _, mdFile := range mdFiles {
		raw, readErr := os.ReadFile(mdFile)
		if readErr != nil {
			return nil, fmt.Errorf("reading %q: %w", mdFile, readErr)
		}
		text := string(raw)
		contentsByFile[mdFile] = text
		anchorsByFile[mdFile] = collectAnchors(text)

		if strings.Contains(text, "dGVzdAo=") {
			rel, _ := filepath.Rel(repoRoot, mdFile)
			findings = append(findings, report.Finding{
				Domain:   domainName,
				Rule:     "stray-base64",
				Severity: report.SeverityError,
				File:     filepath.ToSlash(rel),
				Message:  `contains stray base64 text "dGVzdAo="`,
			})
		}
	}

	for _, mdFile := range mdFiles {
		text := contentsByFile[mdFile]
		relMd, _ := filepath.Rel(repoRoot, mdFile)
		relMd = filepath.ToSlash(relMd)

		for _, m := range mdLinkRe.FindAllStringSubmatch(text, -1) {
			target := strings.TrimSpace(m[2])
			if isExternalTarget(target) {
				continue
			}
			if generatedOutputRe.MatchString(target) {
				continue // Built from templates, not content files.
			}

			pathPart, anchorPart, _ := strings.Cut(target, "#")
			var targetFile string
			switch {
			case pathPart == "":
				targetFile = mdFile
			case strings.HasPrefix(pathPart, "/"):
				// Root-relative links are for generated site files only (handled
				// above). Link to content pages by their source file instead.
				resolved, ok := resolveHugoTarget(contentDir, strings.TrimPrefix(pathPart, "/"))
				if !ok {
					findings = append(findings, report.Finding{
						Domain:   domainName,
						Rule:     "broken-link",
						Severity: report.SeverityError,
						File:     relMd,
						Message:  fmt.Sprintf("broken root-relative link target %q", target),
					})
					continue
				}
				relTarget, _ := filepath.Rel(filepath.Dir(mdFile), resolved)
				suggestion := filepath.ToSlash(relTarget)
				if anchorPart != "" {
					suggestion += "#" + anchorPart
				}
				findings = append(findings, report.Finding{
					Domain:   domainName,
					Rule:     "root-relative-page-link",
					Severity: report.SeverityError,
					File:     relMd,
					Message:  fmt.Sprintf("link %q to its source file instead: %q", target, suggestion),
				})
				// Fall through to the anchor check, so a broken anchor is reported
				// now rather than after the contributor applies the suggestion.
				targetFile = resolved
			case strings.HasSuffix(pathPart, ".md"):
				// File-relative links resolve from the source file's directory,
				// the way hextra's link render hook resolves them.
				candidate := filepath.Clean(filepath.Join(filepath.Dir(mdFile), filepath.FromSlash(pathPart)))
				if !fileExists(candidate) {
					findings = append(findings, report.Finding{
						Domain:   domainName,
						Rule:     "broken-link",
						Severity: report.SeverityError,
						File:     relMd,
						Message:  fmt.Sprintf("broken relative link target %q", target),
					})
					continue
				}
				targetFile = candidate
			default:
				// Page resources, such as images beside the page file, are fine.
				if fileExists(filepath.Join(filepath.Dir(mdFile), filepath.FromSlash(pathPart))) {
					continue
				}
				// Browser-relative links to pages resolve against the rendered URL,
				// which breaks on leaf pages and on GitHub. Require a .md source path.
				findings = append(findings, report.Finding{
					Domain:   domainName,
					Rule:     "url-relative-link",
					Severity: report.SeverityError,
					File:     relMd,
					Message:  fmt.Sprintf("link %q by its source file path ending in .md", target),
				})
				continue
			}

			if anchorPart != "" {
				if anchors, ok := anchorsByFile[targetFile]; ok && !anchors[anchorPart] {
					relTarget, _ := filepath.Rel(repoRoot, targetFile)
					findings = append(findings, report.Finding{
						Domain:   domainName,
						Rule:     "broken-anchor",
						Severity: report.SeverityError,
						File:     relMd,
						Message:  fmt.Sprintf("broken anchor %q in link %q (resolved to %q)", "#"+anchorPart, target, filepath.ToSlash(relTarget)),
					})
				}
			}
		}
	}

	return findings, nil
}

func collectMarkdownFiles(contentDir string) ([]string, error) {
	var mdFiles []string
	err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(mdFiles)
	return mdFiles, nil
}

func collectAnchors(mdText string) map[string]bool {
	anchors := make(map[string]bool)
	for _, m := range headingIDRe.FindAllStringSubmatch(mdText, -1) {
		anchors[m[1]] = true
	}
	for _, m := range headingTextRe.FindAllStringSubmatch(mdText, -1) {
		if slug := slugifyHeading(m[1]); slug != "" {
			anchors[slug] = true
		}
	}
	return anchors
}

func slugifyHeading(text string) string {
	text = shortcodeRe.ReplaceAllString(text, "")
	text = formattingRe.ReplaceAllString(text, "")
	text = strings.ToLower(strings.TrimSpace(text))
	text = nonSlugRe.ReplaceAllString(text, "")
	text = dashSpaceRe.ReplaceAllString(text, "-")
	return strings.Trim(text, "-")
}

// generatedOutputRe matches root-relative links to site files that Hugo
// generates from templates (llms.txt, llms-full*.txt, llms-rules.txt).
var generatedOutputRe = regexp.MustCompile(`^/llms(?:-[a-z0-9-]+)?\.txt$`)

func isExternalTarget(target string) bool {
	return strings.HasPrefix(target, "http://") ||
		strings.HasPrefix(target, "https://") ||
		strings.HasPrefix(target, "mailto:")
}

// resolveHugoTarget maps a Hugo page path, such as "docs/effective-samples"
// or "docs/effective-samples/author-samples", to the content file that
// renders it. Cases are checked in order, and the order sets precedence.
func resolveHugoTarget(baseDir, pathPart string) (string, bool) {
	candidate := filepath.Clean(filepath.Join(baseDir, pathPart))
	switch {
	// A section page: "docs/effective-samples" renders from its _index.md.
	// Checked first so a section wins over a same-named leaf page file.
	case fileExists(filepath.Join(candidate, "_index.md")):
		return filepath.Join(candidate, "_index.md"), true
	// A leaf page: "docs/effective-samples/author-samples" renders from
	// author-samples.md.
	case fileExists(candidate + ".md"):
		return candidate + ".md", true
	// A path that already names a file, such as a link that includes the
	// .md extension or points to a page resource.
	case fileExists(candidate):
		return candidate, true
	default:
		return "", false
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
