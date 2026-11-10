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

package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) failed: %v", fullPath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) failed: %v", fullPath, err)
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		// wantRule is the rule of the first finding, or "" for no findings.
		wantRule string
		// wantMessage, if set, must appear in the first finding's message.
		wantMessage string
		// wantOtherRule, if set, must be the rule of some later finding.
		wantOtherRule string
	}{
		{
			// The supported page-link form: .md source paths resolved from the
			// linking file's directory, including _index.md sections, parent
			// directories (../), anchors, and links from both leaf and section pages.
			name: "valid_file_relative_links",
			files: map[string]string{
				"content/about.md":       "See [Docs](docs/_index.md) and [API](docs/api.md#guidelines).",
				"content/docs/_index.md": "See [Author](author.md#errors) and [About](../about.md).",
				"content/docs/author.md": "## Error Handling {#errors}\nSee [API](api.md#guidelines).",
				"content/docs/api.md":    "### Guidelines {#guidelines}\nSee [Errors](author.md#errors).",
			},
			wantRule: "",
		},
		{
			// A root-relative link that resolves to a content page breaks when
			// browsing the repository; the validator suggests the .md path,
			// keeping the anchor so the suggestion can be applied as-is.
			name: "root_relative_page_link_fails",
			files: map[string]string{
				"content/docs/author.md": "## Error Handling {#errors}",
				"content/docs/api.md":    "See [Errors](/docs/author#errors).",
			},
			wantRule:    "root-relative-page-link",
			wantMessage: `"author.md#errors"`,
		},
		{
			// A root-relative link with a broken anchor reports both problems in
			// one run, so the suggested .md link isn't itself broken.
			name: "root_relative_page_link_with_broken_anchor",
			files: map[string]string{
				"content/docs/author.md": "## Error Handling {#errors}",
				"content/docs/api.md":    "See [Errors](/docs/author#missing-anchor).",
			},
			wantRule:      "root-relative-page-link",
			wantOtherRule: "broken-anchor",
		},
		{
			// A browser-relative link without .md resolves against the rendered URL,
			// which 404s from leaf pages (the bug #50 fixed). Must stay an error.
			name: "url_relative_link_fails",
			files: map[string]string{
				"content/docs/author.md": "## Error Handling {#errors}",
				"content/docs/api.md":    "See [Errors](author#errors).",
			},
			wantRule: "url-relative-link",
		},
		{
			// Relative links to files beside the page, such as images, are page
			// resources, not page links, so they don't need a .md path.
			name: "page_resource_link_allowed",
			files: map[string]string{
				"content/docs/_index.md":   "![Screenshot](example.png)",
				"content/docs/example.png": "png",
			},
			wantRule: "",
		},
		{
			// A .md link whose file doesn't exist relative to the linking file.
			name: "broken_file_relative_link",
			files: map[string]string{
				"content/docs/api.md": "See [Errors](missing.md#errors).",
			},
			wantRule: "broken-link",
		},
		{
			// The target file exists but the anchor doesn't. Uses an anchor that
			// matches neither an explicit {#id} nor a slugified heading.
			name: "broken_file_relative_anchor",
			files: map[string]string{
				"content/docs/author.md": "## Error Handling {#errors}",
				"content/docs/api.md":    "See [Errors](author.md#missing-anchor).",
			},
			wantRule: "broken-anchor",
		},
		{
			// Root-relative links to generated outputs (llms*.txt) have no content
			// file, so they're the one allowed root-relative form.
			name: "generated_llms_links_allowed",
			files: map[string]string{
				"content/docs/guide.md": "See [index](/llms.txt) and [Python](/llms-full-python.txt).\n",
			},
			wantRule: "",
		},
		{
			// A root-relative link that matches no content page and no generated
			// output is reported as broken, not as a style issue.
			name: "broken_root_relative_link",
			files: map[string]string{
				"content/docs/author.md": "See [Result](/docs/style#result).",
			},
			wantRule: "broken-link",
		},
		{
			// A same-page anchor link with no matching heading.
			name: "broken_anchor_link",
			files: map[string]string{
				"content/docs/api.md": "See [Errors](#error-handling).",
			},
			wantRule: "broken-anchor",
		},
		{
			// Guards against a stray placeholder that once shipped in the guide.
			name: "stray_base64_error",
			files: map[string]string{
				"content/docs/author.md": "### Error Handling {#errors}\n\ndGVzdAo=\n",
			},
			wantRule: "stray-base64",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for relPath, content := range tc.files {
				writeTestFile(t, tmpDir, relPath, content)
			}

			findings, err := Check(tmpDir)
			if err != nil {
				t.Fatalf("Check(%q) unexpected error: %v", tmpDir, err)
			}

			if tc.wantRule == "" {
				if len(findings) != 0 {
					t.Errorf("Check(%q) = %v, want 0 findings", tmpDir, findings)
				}
				return
			}

			if len(findings) == 0 {
				t.Fatalf("Check(%q) = 0 findings, want rule %q", tmpDir, tc.wantRule)
			}
			if got := findings[0].Rule; got != tc.wantRule {
				t.Errorf("Check(%q) first finding rule = %q, want %q", tmpDir, got, tc.wantRule)
			}
			if tc.wantMessage != "" && !strings.Contains(findings[0].Message, tc.wantMessage) {
				t.Errorf("Check(%q) first finding message = %q, want it to contain %q", tmpDir, findings[0].Message, tc.wantMessage)
			}
			if tc.wantOtherRule != "" {
				found := false
				for _, f := range findings[1:] {
					if f.Rule == tc.wantOtherRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Check(%q) = %v, want a later finding with rule %q", tmpDir, findings, tc.wantOtherRule)
				}
			}
		})
	}
}
