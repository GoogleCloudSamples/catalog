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

package report

import (
	"bytes"
	"strings"
	"testing"
)

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name     string
		findings []Finding
		want     bool
	}{
		{
			name:     "empty_findings",
			findings: nil,
			want:     false,
		},
		{
			name: "warnings_only",
			findings: []Finding{
				{Domain: "shortcodes", Rule: "unreferenced", Severity: SeverityWarning, File: "a.py", Message: "unused"},
			},
			want: false,
		},
		{
			name: "contains_error",
			findings: []Finding{
				{Domain: "shortcodes", Rule: "unreferenced", Severity: SeverityWarning, File: "a.py", Message: "unused"},
				{Domain: "links", Rule: "broken-anchor", Severity: SeverityError, File: "b.md", Line: 10, Message: "missing"},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := HasErrors(tc.findings); got != tc.want {
				t.Errorf("HasErrors(%v) = %v, want %v", tc.findings, got, tc.want)
			}
		})
	}
}

func TestWriteText(t *testing.T) {
	tests := []struct {
		name       string
		findings   []Finding
		wantSubstr string
	}{
		{
			name:       "no_errors",
			findings:   nil,
			wantSubstr: "All catalog checks passed with 0 errors!",
		},
		{
			name: "error_with_line",
			findings: []Finding{
				{
					Domain:   "snippets",
					Rule:     "unbalanced-quotes",
					Severity: SeverityError,
					File:     "content/docs/code/csharp/result.cs",
					Line:     3,
					Message:  "unbalanced double quote",
				},
			},
			wantSubstr: "[ERROR] [snippets/unbalanced-quotes] content/docs/code/csharp/result.cs:3: unbalanced double quote",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := WriteText(&buf, tc.findings); err != nil {
				t.Fatalf("WriteText(%v) unexpected error: %v", tc.findings, err)
			}
			got := buf.String()
			if !strings.Contains(got, tc.wantSubstr) {
				t.Errorf("WriteText(%v) = %q, want substring %q", tc.findings, got, tc.wantSubstr)
			}
		})
	}
}

func TestWriteGitHub(t *testing.T) {
	tests := []struct {
		name     string
		findings []Finding
		want     string
	}{
		{
			name: "no findings",
			want: "All catalog checks passed with 0 errors!\n",
		},
		{
			name: "error with line and warning without line",
			findings: []Finding{
				{Domain: "links", Rule: "broken-anchor", Severity: SeverityError, File: "content/docs/a.md", Line: 12, Message: "anchor #x not found"},
				{Domain: "snippets", Rule: "author-note", Severity: SeverityWarning, File: "content/docs/code/go/b.go", Message: "TODO found"},
			},
			want: "::error file=content/docs/a.md,line=12,title=links/broken-anchor::anchor #x not found\n" +
				"::warning file=content/docs/code/go/b.go,title=snippets/author-note::TODO found\n" +
				"Catalog checks found 1 error(s) and 1 warning(s).\n",
		},
		{
			name: "escapes special characters",
			findings: []Finding{
				{Domain: "d", Rule: "r", Severity: SeverityError, File: "a,b:c.md", Message: "100% bad\nnext line"},
			},
			want: "::error file=a%2Cb%3Ac.md,title=d/r::100%25 bad%0Anext line\n" +
				"Catalog checks found 1 error(s) and 0 warning(s).\n",
		},
		{
			name: "finding without file",
			findings: []Finding{
				{Domain: "languages", Rule: "missing", Severity: SeverityError, Message: "no output format"},
			},
			want: "::error title=languages/missing::no output format\n" +
				"Catalog checks found 1 error(s) and 0 warning(s).\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			if err := WriteGitHub(&b, tc.findings); err != nil {
				t.Fatalf("WriteGitHub() error = %v", err)
			}
			if got := b.String(); got != tc.want {
				t.Errorf("WriteGitHub() =\n%q\nwant\n%q", got, tc.want)
			}
		})
	}
}
