package snippets

import (
	"os"
	"path/filepath"
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
		name     string
		relPath  string
		content  string
		wantRule string
	}{
		{
			name:     "valid_snippet_with_region_tags",
			relPath:  "content/docs/code/go/region-tags.go",
			content:  "// [START sandwichmanager_make_sandwich]\nfunc example() {}\n// [END sandwichmanager_make_sandwich]\n",
			wantRule: "",
		},
		{
			name:     "trailing_comment_and_char_literal_quotes_ignored",
			relPath:  "content/docs/code/go/example.go",
			content:  "func example() {\n\tch := '\"'\n\tx := 1 // says \"hello\n\traw := `line 1 \"quoted\nline 2`\n\t_ = ch\n\t_ = x\n\t_ = raw\n}\n",
			wantRule: "",
		},
		{
			name:     "multiline_triple_quotes_ignored",
			relPath:  "content/docs/code/java/example.java",
			content:  "class Example {\n  String s = \"\"\"\n    hello \"world\n    \"\"\";\n}\n",
			wantRule: "",
		},
		{
			name:     "leaked_author_note",
			relPath:  "content/docs/code/python/clients.py",
			content:  "# NOTE FOR SAMPLE AUTHOR: remove this\nwith Client() as c:\n    pass\n",
			wantRule: "leaked-author-note",
		},
		{
			name:     "unbalanced_double_quote",
			relPath:  "content/docs/code/csharp/result.cs",
			content:  "public void Example(string projectId = \"my-project-id) {}\n",
			wantRule: "unbalanced-quotes",
		},
		{
			name:     "invalid_region_tag_format",
			relPath:  "content/docs/code/go/region-tags.go",
			content:  "// [START Product-BadTag]\nfunc example() {}\n// [END Product-BadTag]\n",
			wantRule: "region-tag-format",
		},
		{
			name:     "unmatched_region_tag",
			relPath:  "content/docs/code/go/region-tags.go",
			content:  "// [START sandwichmanager_make_sandwich]\nfunc example() {}\n",
			wantRule: "region-tag-mismatch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			writeTestFile(t, tmpDir, tc.relPath, tc.content)

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
		})
	}
}

func TestCheckRegionTagsDeterministicOrder(t *testing.T) {
	tmpDir := t.TempDir()
	content := "// [START product_zeta]\n// [START product_alpha]\nfunc example() {}\n"
	writeTestFile(t, tmpDir, "content/docs/code/go/region-tags.go", content)

	findings, err := Check(tmpDir)
	if err != nil {
		t.Fatalf("Check(%q) unexpected error: %v", tmpDir, err)
	}
	if len(findings) != 2 {
		t.Fatalf("Check(%q) returned %d findings, want 2", tmpDir, len(findings))
	}
	if findings[0].Line != 1 || findings[1].Line != 2 {
		t.Errorf("Check(%q) finding lines = (%d, %d), want (1, 2)", tmpDir, findings[0].Line, findings[1].Line)
	}
}
