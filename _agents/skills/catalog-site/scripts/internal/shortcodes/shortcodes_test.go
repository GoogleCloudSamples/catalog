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

package shortcodes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/report"
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
		name     string
		files    map[string]string
		wantRule string
		wantSev  report.Severity
	}{
		{
			name: "valid_lowercase_match",
			files: map[string]string{
				"content/docs/effective-samples/author.md": `{{% guideline_examples "imports" %}}`,
				"content/docs/code/python/imports.py":      "import os\n",
				"content/docs/code/python/imports.md":      "Include imports inside region tags.\n",
				"content/docs/code/python/.DS_Store":       "",
				"content/docs/code/python/README.md":       "# Notes\n",
			},
			wantRule: "",
		},
		{
			name: "case_mismatch_error",
			files: map[string]string{
				"content/docs/effective-samples/author.md": `{{% guideline_examples "imports" %}}`,
				"content/docs/code/csharp/Imports.cs":      "using System;\n",
			},
			wantRule: "case-mismatch",
			wantSev:  report.SeverityError,
		},
		{
			name: "na_stub_file_error",
			files: map[string]string{
				"content/docs/effective-samples/author.md": `{{% guideline_examples "clients" %}}`,
				"content/docs/code/terraform/clients.tf":   "# N/A\n",
			},
			wantRule: "na-stub-file",
			wantSev:  report.SeverityError,
		},
		{
			name: "unknown_tab_name_warning",
			files: map[string]string{
				"content/docs/effective-samples/author.md": "{{< tabs >}}\n{{< tab name=\"NodeJS\" >}}x{{< /tab >}}\n{{< /tabs >}}\n",
			},
			wantRule: "unknown-tab-name",
			wantSev:  report.SeverityWarning,
		},
		{
			name: "unreferenced_guideline_warning",
			files: map[string]string{
				"content/docs/effective-samples/author.md": `{{% guideline_examples "imports" %}}`,
				"content/docs/code/python/imports.py":      "import os\n",
				"content/docs/code/python/errors.py":       "try:\n    pass\nexcept Exception:\n    pass\n",
			},
			wantRule: "unreferenced-guideline",
			wantSev:  report.SeverityWarning,
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
			got := findings[0]
			if got.Rule != tc.wantRule || got.Severity != tc.wantSev {
				t.Errorf("Check(%q) first finding = (%q, %q), want (%q, %q)", tmpDir, got.Rule, got.Severity, tc.wantRule, tc.wantSev)
			}
		})
	}
}
