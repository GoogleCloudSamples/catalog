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

// Package templates statically checks the Hugo layouts that render the agent
// and LLM text outputs (<page>.md, llms-full*.txt, llms-rules.txt), without
// building the site.
package templates

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

const domainName = "templates"

var (
	shortcodeRe = regexp.MustCompile(`\{\{[<%]\s*([a-z_]+)`)
	commentRe   = regexp.MustCompile(`(?s)\{\{-?\s*/\*.*?\*/\s*-?\}\}`)
	actionRe    = regexp.MustCompile(`(?s)\{\{-?(.*?)-?\}\}`)
	// controlRe matches actions that do not print a value.
	controlRe = regexp.MustCompile(`^(if|else|end|range|with|define|template|block|return|partial|break|continue)\b|^\$[A-Za-z_]+\s*:?=`)
)

// Check validates the layouts under repoRoot/layouts that render the agent and
// LLM text outputs.
func Check(repoRoot string) ([]report.Finding, error) {
	c := &checker{root: repoRoot}
	if err := c.homeTemplates(); err != nil {
		return nil, err
	}
	if err := c.unescapedValues(); err != nil {
		return nil, err
	}
	if err := c.textShortcodes(); err != nil {
		return nil, err
	}
	return c.findings, nil
}

type checker struct {
	root     string
	findings []report.Finding
}

func (c *checker) add(rule, file string, line int, format string, args ...any) {
	c.findings = append(c.findings, report.Finding{
		Domain:   domainName,
		Rule:     rule,
		Severity: report.SeverityError,
		File:     file,
		Line:     line,
		Message:  fmt.Sprintf(format, args...),
	})
}

func (c *checker) exists(rel string) bool {
	info, err := os.Stat(filepath.Join(c.root, filepath.FromSlash(rel)))
	return err == nil && !info.IsDir()
}

// homeTemplates flags layouts/home.llms*.txt. Hugo prefers a home template
// for another text output format over the exact-format llms.txt, so such a
// file replaces the /llms.txt index.
func (c *checker) homeTemplates() error {
	matches, err := filepath.Glob(filepath.Join(c.root, "layouts", "home.llms*.txt"))
	if err != nil {
		return fmt.Errorf("listing layouts: %w", err)
	}
	for _, m := range matches {
		name := filepath.Base(m)
		c.add("home-llms-template", "layouts/"+name, 0,
			"rename to layouts/%s; a home.llms*.txt layout replaces the /llms.txt index", strings.TrimPrefix(name, "home."))
	}
	return nil
}

// unescapedValues flags printed values in layouts/_partials/llms/ that do not
// use safeHTML. Hugo parses partials as HTML templates, so unmarked values are
// HTML-escaped in the plain-text outputs.
func (c *checker) unescapedValues() error {
	dir := filepath.Join(c.root, "layouts", "_partials", "llms")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // Not configured.
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		rel := "layouts/_partials/llms/" + e.Name()
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return fmt.Errorf("reading %q: %w", rel, err)
		}
		src := commentRe.ReplaceAllStringFunc(string(raw), func(s string) string {
			return strings.Repeat("\n", strings.Count(s, "\n"))
		})
		for _, loc := range actionRe.FindAllStringSubmatchIndex(src, -1) {
			body := strings.TrimSpace(src[loc[2]:loc[3]])
			if body == "" || controlRe.MatchString(body) || strings.Contains(body, "safeHTML") {
				continue
			}
			line := strings.Count(src[:loc[0]], "\n") + 1
			c.add("unescaped-partial-value", rel, line,
				"printed value %q is HTML-escaped in text outputs; pipe it through safeHTML", body)
		}
	}
	return nil
}

// textShortcodes requires a .md and .txt override for every shortcode used in
// content/, so shortcode HTML does not leak into the text outputs.
func (c *checker) textShortcodes() error {
	if !c.exists("layouts/page.markdown.md") {
		return nil // Text outputs not configured.
	}
	contentDir := filepath.Join(c.root, "content")
	codeDir := filepath.Join(contentDir, "docs", "code")
	used := map[string]string{} // Shortcode name to first file using it.
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
		if path == filepath.Join(contentDir, "_index.md") {
			return nil // The home page has no Markdown or llms-full output.
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %q: %w", path, err)
		}
		rel, _ := filepath.Rel(c.root, path)
		for _, m := range shortcodeRe.FindAllStringSubmatch(string(raw), -1) {
			if _, ok := used[m[1]]; !ok {
				used[m[1]] = filepath.ToSlash(rel)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("scanning content shortcodes: %w", err)
	}
	names := make([]string, 0, len(used))
	for name := range used {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		for _, ext := range []string{"md", "txt"} {
			rel := fmt.Sprintf("layouts/_shortcodes/%s.%s", name, ext)
			if !c.exists(rel) {
				c.add("missing-text-shortcode", rel, 0,
					"shortcode %q (used in %s) has no .%s override; its HTML would leak into the text outputs", name, used[name], ext)
			}
		}
	}
	return nil
}
