package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) failed: %v", p, err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) failed: %v", p, err)
	}
}

// validFiles returns a minimal repository with complete text outputs for one
// language (python) and one content shortcode (callout).
func validFiles() map[string]string {
	return map[string]string{
		"hugo.yaml": "outputs:\n  home: [html, llms, llmsfull, llmspython]\n" +
			"outputFormats:\n  llmsfull:\n    baseName: llms-full\n  llmspython:\n    baseName: llms-full-python\n",
		"layouts/_partials/samples/languages.html":              `(dict "name" "Python" "lang" "python" "ext" "py" "folder" "python")`,
		"layouts/page.markdown.md":                              "# {{ .Title }}\n",
		"layouts/llmsfull.txt":                                  "{{- partial \"llms/full.txt\" (dict) -}}\n",
		"layouts/llmspython.txt":                                "{{- partial \"llms/full.txt\" (dict) -}}\n",
		"layouts/_shortcodes/guideline_examples.llmspython.txt": "{{- partial \"llms/examples.txt\" (dict) -}}\n",
		"layouts/_shortcodes/tab.llmspython.txt":                "{{- partial \"llms/tab.txt\" (dict) -}}\n",
		"layouts/_shortcodes/callout.md":                        "{{- partial \"llms/callout.txt\" (dict) -}}\n",
		"layouts/_shortcodes/callout.txt":                       "{{- partial \"llms/callout.txt\" (dict) -}}\n",
		"layouts/_partials/llms/callout.txt": "{{- /* Callout.\n  Params: type. */ -}}\n" +
			"{{- $label := .type | title -}}\n> **{{ $label | safeHTML }}:** {{ .inner | safeHTML }}\n{{ if .x }}{{ end }}\n",
		"content/docs/guide.md": "{{< callout >}}Note{{< /callout >}}\n",
		"content/_index.md":     "{{< cards >}}{{< card link=\"docs\" >}}{{< /cards >}}\n",
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(map[string]string)
		wantRule string
		wantText string
	}{
		{name: "Valid", mutate: func(map[string]string) {}},
		{
			name: "HomeTemplateReplacesIndex",
			mutate: func(f map[string]string) {
				f["layouts/home.llmscsharp.txt"] = "x"
			},
			wantRule: "home-llms-template",
		},
		{
			name: "UnescapedPartialValue",
			mutate: func(f map[string]string) {
				f["layouts/_partials/llms/callout.txt"] += "{{ os.ReadFile $code | strings.TrimSpace }}\n"
			},
			wantRule: "unescaped-partial-value",
			wantText: "os.ReadFile",
		},
		{
			name: "MissingTextShortcode",
			mutate: func(f map[string]string) {
				f["content/docs/guide.md"] += "{{% details title=\"x\" %}}y{{% /details %}}\n"
			},
			wantRule: "missing-text-shortcode",
			wantText: "details",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			files := validFiles()
			tc.mutate(files)
			for rel, content := range files {
				writeFile(t, root, rel, content)
			}
			findings, err := Check(root)
			if err != nil {
				t.Fatalf("Check() unexpected error: %v", err)
			}
			if tc.wantRule == "" {
				if len(findings) != 0 {
					t.Errorf("Check() = %v, want no findings", findings)
				}
				return
			}
			found := false
			for _, f := range findings {
				if f.Rule == tc.wantRule && strings.Contains(f.File+f.Message, tc.wantText) {
					found = true
				}
			}
			if !found {
				t.Errorf("Check() = %v, want rule %q mentioning %q", findings, tc.wantRule, tc.wantText)
			}
		})
	}
}
