package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// WriteJSON serializes the monthly metrics report as indented JSON.
func WriteJSON(w io.Writer, r *Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("encoding metrics json: %w", err)
	}
	return nil
}

// WriteText writes a concise human-readable summary of the monthly sample
// metrics to w.
func WriteText(w io.Writer, r *Report) error {
	if _, err := fmt.Fprintf(
		w,
		"Sample Metrics Report (%s: %s to %s)\nCreated: %d samples | Updated: %d samples | Languages: %s\n\n",
		r.MonthLabel,
		r.Since.Format("2006-01-02"),
		r.Until.Format("2006-01-02"),
		r.TotalCreated,
		r.TotalUpdated,
		strings.Join(r.Languages, ", "),
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "By Repository:"); err != nil {
		return err
	}
	for _, repo := range r.Repos {
		if _, err := fmt.Fprintf(
			w,
			"  - %-45s created=%-3d updated=%-3d (commits=%d, skipped_bot_or_deps=%d)\n",
			repo.Repo,
			repo.CreatedSamples,
			repo.UpdatedSamples,
			repo.TotalCommits,
			repo.SkippedBotOrDepCommits,
		); err != nil {
			return err
		}
	}

	if len(r.Products) > 0 {
		if _, err := fmt.Fprintln(w, "\nBy Product:"); err != nil {
			return err
		}
		for _, p := range r.Products {
			if _, err := fmt.Fprintf(
				w,
				"  - %-35s created=%-3d updated=%-3d [%s]\n",
				p.Product,
				len(p.Created),
				len(p.Updated),
				strings.Join(p.Languages, ", "),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteMarkdown renders the monthly metrics report in the Release Notes
// format matching archetypes/blog.md, followed by a context digest of commit
// subjects and sample links to assist with blog authoring.
func WriteMarkdown(w io.Writer, r *Report) error {
	var b strings.Builder

	fmt.Fprintf(&b, "## Release Notes\n\n")
	fmt.Fprintf(
		&b,
		"This month we added **%d new samples** and updated **%d existing samples** across **%d languages** (%s) in the Samples Catalog.\n\n",
		r.TotalCreated,
		r.TotalUpdated,
		len(r.Languages),
		strings.Join(r.Languages, ", "),
	)

	var hasCreated, hasUpdated bool
	for _, p := range r.Products {
		if len(p.Created) > 0 {
			hasCreated = true
		}
		if len(p.Updated) > 0 {
			hasUpdated = true
		}
	}

	if hasCreated {
		b.WriteString("### New Sample Coverage\n\n")
		for _, p := range r.Products {
			if len(p.Created) == 0 {
				continue
			}
			langs := uniqueLanguages(p.Created)
			subjects := uniqueCommitSubjects(p.Created, 3)
			example := p.Created[0]
			fmt.Fprintf(
				&b,
				"* **%s** (%s — %d new %s): %s ([`%s`](%s))\n",
				p.Product,
				strings.Join(langs, ", "),
				len(p.Created),
				pluralize("sample", len(p.Created)),
				strings.Join(subjects, "; "),
				displaySampleLabel(example.FilePath),
				example.HTMLURL,
			)
		}
		b.WriteString("\n")
	}

	if hasUpdated {
		b.WriteString("### Sample Updates & Modernization\n\n")
		for _, p := range r.Products {
			if len(p.Updated) == 0 {
				continue
			}
			langs := uniqueLanguages(p.Updated)
			subjects := uniqueCommitSubjects(p.Updated, 3)
			example := p.Updated[0]
			fmt.Fprintf(
				&b,
				"* **%s** (%s — %d updated %s): %s ([`%s`](%s))\n",
				p.Product,
				strings.Join(langs, ", "),
				len(p.Updated),
				pluralize("sample", len(p.Updated)),
				strings.Join(subjects, "; "),
				displaySampleLabel(example.FilePath),
				example.HTMLURL,
			)
		}
		b.WriteString("\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func uniqueLanguages(activities []SampleActivity) []string {
	var langs []string
	for _, a := range activities {
		langs = appendUnique(langs, a.Language)
	}
	sort.Strings(langs)
	return langs
}

func uniqueCommitSubjects(activities []SampleActivity, limit int) []string {
	var subjects []string
	for _, a := range activities {
		for _, c := range a.Commits {
			clean := cleanCommitSubject(c.Subject)
			if clean != "" {
				subjects = appendUnique(subjects, clean)
			}
			if limit > 0 && len(subjects) >= limit {
				return subjects
			}
		}
	}
	return subjects
}

var conventionalPrefixRe = regexp.MustCompile(`^(?:feat|fix|docs|chore|refactor|test|style|perf)(?:\([^)]+\))?!?:\s*`)

func cleanCommitSubject(subject string) string {
	s := strings.TrimSpace(subject)
	s = conventionalPrefixRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

var genericFilenames = map[string]bool{
	"app.go":    true,
	"app.js":    true,
	"app.py":    true,
	"app.rb":    true,
	"client.go": true,
	"index.js":  true,
	"index.php": true,
	"index.ts":  true,
	"main.go":   true,
	"main.py":   true,
	"main.tf":   true,
	"server.go": true,
	"server.js": true,
}

func displaySampleLabel(filePath string) string {
	base := filepath.Base(filePath)
	if genericFilenames[strings.ToLower(base)] {
		return filepath.ToSlash(filePath)
	}
	return base
}
