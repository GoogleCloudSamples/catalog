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

package languages

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

// validFiles returns a repository whose language list matches
// report.Languages, with every language's output format and templates.
func validFiles() map[string]string {
	folders := []string{"python", "node", "java", "go", "csharp", "ruby", "php", "terraform"}
	formats := []string{"llmsfull"}
	for _, f := range folders {
		formats = append(formats, "llms"+f)
	}
	cfg := "outputs:\n  home: [html, llms, " + strings.Join(formats, ", ") + "]\noutputFormats:\n"
	files := map[string]string{
		"layouts/_partials/samples/languages.html": "LANGS",
		"layouts/llmsfull.txt":                     "x",
	}
	for _, f := range formats {
		cfg += "  " + f + ":\n    baseName: x\n"
	}
	for _, f := range folders {
		files["layouts/llms"+f+".txt"] = "x"
		files["layouts/_shortcodes/guideline_examples.llms"+f+".txt"] = "x"
		files["layouts/_shortcodes/tab.llms"+f+".txt"] = "x"
	}
	files["hugo.yaml"] = cfg
	return files
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
			name: "LanguageListDrift",
			mutate: func(f map[string]string) {
				f["layouts/_partials/samples/languages.html"] = `(dict "name" "Python" "lang" "python" "ext" "py" "folder" "python")`
			},
			wantRule: "language-list-drift",
		},
		{
			name: "MissingOutputFormat",
			mutate: func(f map[string]string) {
				f["hugo.yaml"] = strings.Replace(f["hugo.yaml"], "  llmsgo:\n", "", 1)
			},
			wantRule: "incomplete-language-outputs",
			wantText: "outputFormats.llmsgo",
		},
		{
			name: "MissingHomeOutput",
			mutate: func(f map[string]string) {
				f["hugo.yaml"] = strings.Replace(f["hugo.yaml"], ", llmsruby", "", 1)
			},
			wantRule: "incomplete-language-outputs",
			wantText: "outputs.home does not include llmsruby",
		},
		{
			name: "MissingTemplate",
			mutate: func(f map[string]string) {
				delete(f, "layouts/_shortcodes/tab.llmspython.txt")
			},
			wantRule: "incomplete-language-outputs",
			wantText: "tab.llmspython.txt",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			files := validFiles()
			files["layouts/_partials/samples/languages.html"] = "(dict \"name\" \"Python\" \"lang\" \"x\" \"ext\" \"py\" \"folder\" \"python\")\n(dict \"name\" \"Node.js\" \"lang\" \"x\" \"ext\" \"js\" \"folder\" \"node\")\n(dict \"name\" \"Java\" \"lang\" \"x\" \"ext\" \"java\" \"folder\" \"java\")\n(dict \"name\" \"Go\" \"lang\" \"x\" \"ext\" \"go\" \"folder\" \"go\")\n(dict \"name\" \"C#\" \"lang\" \"x\" \"ext\" \"cs\" \"folder\" \"csharp\")\n(dict \"name\" \"Ruby\" \"lang\" \"x\" \"ext\" \"rb\" \"folder\" \"ruby\")\n(dict \"name\" \"PHP\" \"lang\" \"x\" \"ext\" \"php\" \"folder\" \"php\")\n(dict \"name\" \"Terraform\" \"lang\" \"x\" \"ext\" \"tf\" \"folder\" \"terraform\")"
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
				if f.Rule == tc.wantRule && strings.Contains(f.File+" "+f.Message, tc.wantText) {
					found = true
				}
			}
			if !found {
				t.Errorf("Check() = %v, want rule %q mentioning %q", findings, tc.wantRule, tc.wantText)
			}
		})
	}
}
