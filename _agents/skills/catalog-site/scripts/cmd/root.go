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

// Package cmd wires the Cobra CLI commands for the catalog validation tool.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/languages"
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/links"
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/report"
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/shortcodes"
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/snippets"
	"github.com/GoogleCloudSamples/catalog/skills/catalog-site/scripts/internal/templates"
	"github.com/spf13/cobra"
)

// ErrValidationFailed indicates that one or more blocking validation errors
// were reported.
var ErrValidationFailed = errors.New("catalog validation failed")

type options struct {
	repoRoot string
	format   string
}

type checkFunc func(repoRoot string) ([]report.Finding, error)

// NewRootCmd constructs the root Cobra command and attaches all domain
// subcommands without relying on mutable package-level global state.
func NewRootCmd() *cobra.Command {
	opts := &options{}

	rootCmd := &cobra.Command{
		Use:           "validate_catalog",
		Short:         "Validate the Cloud Samples Catalog Hugo repository",
		Long:          "Runs shortcode, link/anchor, multi-language snippet, agent/LLM template, and language setup validations across the Cloud Samples Catalog repository.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runChecks(cmd.OutOrStdout(), opts, shortcodes.Check, links.Check, snippets.Check, templates.Check, languages.Check)
		},
	}

	defaultRoot := defaultRepoRoot()
	rootCmd.PersistentFlags().StringVar(&opts.repoRoot, "repo-root", defaultRoot, "Path to the catalog repository root")
	rootCmd.PersistentFlags().StringVar(&opts.format, "format", "text", "Output format (text, json, or github)")

	rootCmd.AddCommand(
		newShortcodesCmd(opts),
		newLinksCmd(opts),
		newSnippetsCmd(opts),
		newTemplatesCmd(opts),
		newLanguagesCmd(opts),
		newMetricsCmd(),
	)

	return rootCmd
}

func runChecks(w io.Writer, opts *options, checks ...checkFunc) error {
	var allFindings []report.Finding
	for _, check := range checks {
		findings, err := check(opts.repoRoot)
		if err != nil {
			return err
		}
		allFindings = append(allFindings, findings...)
	}

	switch opts.format {
	case "text":
		if err := report.WriteText(w, allFindings); err != nil {
			return err
		}
	case "json":
		if err := report.WriteJSON(w, allFindings); err != nil {
			return err
		}
	case "github":
		if err := report.WriteGitHub(w, allFindings); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported output format %q (want \"text\", \"json\", or \"github\")", opts.format)
	}

	if report.HasErrors(allFindings) {
		return ErrValidationFailed
	}
	return nil
}

func defaultRepoRoot() string {
	if wd, err := os.Getwd(); err == nil {
		for dir := wd; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
			if _, statErr := os.Stat(filepath.Join(dir, "hugo.yaml")); statErr == nil {
				return dir
			}
		}
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	// _agents/skills/catalog-site/scripts/cmd/root.go -> 5 levels up is repo root.
	dir := filepath.Dir(filename)
	for i := 0; i < 5; i++ {
		dir = filepath.Dir(dir)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return abs
}
