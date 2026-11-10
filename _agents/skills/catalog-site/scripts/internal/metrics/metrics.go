// Package metrics collects monthly created and updated sample metrics across
// the Google Cloud sample repositories via the GitHub REST API.
package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// DefaultRepos lists the canonical Google Cloud sample repositories queried by
// default when collecting monthly catalog metrics.
var DefaultRepos = []string{
	"GoogleCloudSamples/client-library-samples",
	"GoogleCloudPlatform/python-docs-samples",
	"GoogleCloudPlatform/nodejs-docs-samples",
	"GoogleCloudPlatform/java-docs-samples",
	"GoogleCloudPlatform/golang-samples",
	"GoogleCloudPlatform/dotnet-docs-samples",
	"GoogleCloudPlatform/php-docs-samples",
	"GoogleCloudPlatform/ruby-docs-samples",
	"terraform-google-modules/terraform-docs-samples",
}

// ChangeType classifies whether a sample was newly created or updated during
// the reporting window.
type ChangeType string

const (
	// ChangeCreated indicates a sample that was newly introduced during the window.
	ChangeCreated ChangeType = "created"
	// ChangeUpdated indicates an existing sample whose code was modified during the window.
	ChangeUpdated ChangeType = "updated"
)

// Options configures a monthly metrics collection run.
type Options struct {
	Since      time.Time
	Until      time.Time
	MonthLabel string
	Repos      []string
	Token      string
	BaseURL    string
	HTTPClient *http.Client
}

// CommitRef summarizes a commit that touched a sample.
type CommitRef struct {
	SHA     string    `json:"sha"`
	Subject string    `json:"subject"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	HTMLURL string    `json:"html_url"`
}

// SampleActivity records the net activity for a single sample file in a window.
type SampleActivity struct {
	Repo       string      `json:"repo"`
	FilePath   string      `json:"file_path"`
	Product    string      `json:"product"`
	Language   string      `json:"language"`
	Change     ChangeType  `json:"change"`
	RegionTags []string    `json:"region_tags,omitempty"`
	HTMLURL    string      `json:"html_url"`
	Commits    []CommitRef `json:"commits"`
}

// ProductSummary aggregates created and updated samples for a product area.
type ProductSummary struct {
	Product   string           `json:"product"`
	Created   []SampleActivity `json:"created,omitempty"`
	Updated   []SampleActivity `json:"updated,omitempty"`
	Languages []string         `json:"languages"`
}

// RepoSummary aggregates activity metrics for a single GitHub repository.
type RepoSummary struct {
	Repo                   string `json:"repo"`
	TotalCommits           int    `json:"total_commits"`
	SkippedBotOrDepCommits int    `json:"skipped_bot_or_dep_commits"`
	CreatedSamples         int    `json:"created_samples"`
	UpdatedSamples         int    `json:"updated_samples"`
}

// Report contains the complete monthly sample activity metrics across all
// target repositories.
type Report struct {
	MonthLabel   string           `json:"month_label"`
	Since        time.Time        `json:"since"`
	Until        time.Time        `json:"until"`
	TotalCreated int              `json:"total_created"`
	TotalUpdated int              `json:"total_updated"`
	Languages    []string         `json:"languages"`
	Repos        []RepoSummary    `json:"repos"`
	Products     []ProductSummary `json:"products"`
}

var (
	regionTagStartRe = regexp.MustCompile(`\[START\s+([a-zA-Z0-9_.-]+)\]`)
	regionTagEndRe   = regexp.MustCompile(`\[END\s+([a-zA-Z0-9_.-]+)\]`)
)

var ignoredWrapperTags = map[string]bool{
	"all":      true,
	"def":      true,
	"header":   true,
	"import":   true,
	"imports":  true,
	"main":     true,
	"requires": true,
}

var productDisplayNames = map[string]string{
	"aiplatform":            "Gemini Platform",
	"AlloyDB":               "AlloyDB",
	"alloydb":               "AlloyDB",
	"appengine":             "App Engine",
	"artifactregistry":      "Artifact Registry",
	"asset":                 "Cloud Asset Inventory",
	"auth":                  "Authentication",
	"automl":                "AutoML",
	"backupdr":              "Backup and DR",
	"batch":                 "Batch",
	"bigquery":              "BigQuery",
	"bigqueryconnection":    "BigQuery Connection",
	"bigquerydatapolicy":    "BigQuery Data Policy",
	"bigquerydatatransfer":  "BigQuery Data Transfer",
	"bigquerymigration":     "BigQuery Migration",
	"bigqueryreservation":   "BigQuery Reservation",
	"bigquerystorage":       "BigQuery Storage",
	"bigtable":              "Cloud Bigtable",
	"cdn":                   "Cloud CDN",
	"cloud-sql":             "Cloud SQL",
	"cloud_sql":             "Cloud SQL",
	"cloudbuild":            "Cloud Build",
	"cloudrun":              "Cloud Run",
	"cloudsql":              "Cloud SQL",
	"composer":              "Cloud Composer",
	"compute":               "Compute Engine",
	"confidentialcomputing": "Confidential Computing",
	"container":             "Google Kubernetes Engine",
	"datacatalog":           "Data Catalog",
	"dataflow":              "Dataflow",
	"dataform":              "Dataform",
	"datafusion":            "Cloud Data Fusion",
	"datalabeling":          "Data Labeling",
	"dataplex":              "Dataplex",
	"dataproc":              "Dataproc",
	"datastore":             "Firestore in Datastore Mode",
	"deploy":                "Cloud Deploy",
	"developer-knowledge":   "Developer Knowledge",
	"developerknowledge":    "Developer Knowledge",
	"dialogflow":            "Dialogflow",
	"dialogflow-cx":         "Dialogflow CX",
	"dlp":                   "Sensitive Data Protection (DLP)",
	"kms":                   "Cloud KMS",
	"dns":                   "Cloud DNS",
	"documentai":            "Document AI",
	"endpoints":             "Cloud Endpoints",
	"error_reporting":       "Error Reporting",
	"eventarc":              "Eventarc",
	"filestore":             "Filestore",
	"firestore":             "Firestore",
	"functions":             "Cloud Run Functions",
	"genai":                 "Generative AI",
	"generative_ai":         "Vertex AI Generative AI",
	"gke":                   "Google Kubernetes Engine",
	"gkehub":                "GKE Enterprise",
	" healthcare":           "Cloud Healthcare API",
	"healthcare":            "Cloud Healthcare API",
	"iam":                   "Identity and Access Management (IAM)",
	"iap":                   "Identity-Aware Proxy (IAP)",
	"iot":                   "Cloud IoT Core",
	"jobs":                  "Cloud Talent Solution",
	"kubernetes_engine":     "Google Kubernetes Engine",
	"language":              "Cloud Natural Language",
	"lb":                    "Cloud Load Balancing",
	"logging":               "Cloud Logging",
	"media":                 "Media CDN / Transcoder",
	"memorystore":           "Memorystore",
	" model_armor":          "Model Armor",
	"model_armor":           "Model Armor",
	"monitoring":            "Cloud Monitoring",
	"network_connectivity":  "Network Connectivity",
	"network_management":    "Network Intelligence Center",
	"networkmanagement":     "Network Intelligence Center",
	"opentelemetry":         "Cloud Trace / OpenTelemetry",
	"OR":                    "Operations",
	"optimization":          "Cloud Fleet Routing",
	" parameterized":        "Parameterized",
	"people-and-planet-ai":  "People and Planet AI",
	"privateca":             "Certificate Authority Service",
	"profiler":              "Cloud Profiler",
	"pubsub":                "Pub/Sub",
	"pubsublite":            "Pub/Sub Lite",
	"recaptcha_enterprise":  "reCAPTCHA Enterprise",
	"retail":                "Vertex AI Search for Commerce",
	"routeoptimization":     "Route Optimization",
	"run":                   "Cloud Run",
	"scheduler":             "Cloud Scheduler",
	"secretmanager":         "Secret Manager",
	"security-center":       "Security Command Center",
	"securitycenter":        "Security Command Center",
	"servicedirectory":      "Service Directory",
	"spanner":               "Spanner",
	"speech":                "Cloud Speech-to-Text",
	"storage":               "Cloud Storage",
	"storage-control":       "Cloud Storage Control",
	"storage_control":       "Cloud Storage Control",
	"storagecontrol":        "Cloud Storage Control",
	"storageinsights":       "Storage Insights",
	"storagetransfer":       "Storage Transfer Service",
	"tables":                "AutoML Tables",
	"tasks":                 "Cloud Tasks",
	"texttospeech":          "Cloud Text-to-Speech",
	"trace":                 "Cloud Trace",
	"translate":             "Cloud Translation",
	"vertexai":              "Vertex AI",
	"video":                 "Cloud Video Intelligence",
	"videointelligence":     "Cloud Video Intelligence",
	"vision":                "Cloud Vision",
	"vmwareengine":          "Google Cloud VMware Engine",
	"vpcaccess":             "Serverless VPC Access",
	"workflows":             "Workflows",
	"workstations":          "Cloud Workstations",
}

// ResolveGitHubToken finds a GitHub API token from environment variables or the
// GitHub CLI (`gh auth token`). It returns an empty string if none is available.
func ResolveGitHubToken(ctx context.Context) string {
	for _, env := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if tok := strings.TrimSpace(os.Getenv(env)); tok != "" {
			return tok
		}
	}
	cmd := exec.CommandContext(ctx, "gh", "auth", "token")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ParseMonthWindow parses a YYYY-MM string into the UTC start and end instants
// of that calendar month.
func ParseMonthWindow(month string) (time.Time, time.Time, error) {
	t, err := time.Parse("2006-01", strings.TrimSpace(month))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parsing month %q (want YYYY-MM): %w", month, err)
	}
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	return start, end, nil
}

// Collect queries the GitHub API for all configured repositories in the time
// window and returns the aggregated monthly sample metrics report.
func Collect(ctx context.Context, opts Options) (*Report, error) {
	repos := opts.Repos
	if len(repos) == 0 {
		repos = DefaultRepos
	}
	baseURL := strings.TrimRight(opts.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	client := &ghClient{
		baseURL:    baseURL,
		token:      opts.Token,
		httpClient: httpClient,
	}

	type repoResult struct {
		index      int
		summary    RepoSummary
		activities []SampleActivity
		err        error
	}

	results := make([]repoResult, len(repos))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)

	for i, repo := range repos {
		wg.Add(1)
		go func(idx int, repoName string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			summary, activities, err := collectRepo(ctx, client, repoName, opts.Since, opts.Until)
			results[idx] = repoResult{
				index:      idx,
				summary:    summary,
				activities: activities,
				err:        err,
			}
		}(i, repo)
	}
	wg.Wait()

	var (
		repoSummaries []RepoSummary
		allActivities []SampleActivity
	)
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		repoSummaries = append(repoSummaries, r.summary)
		allActivities = append(allActivities, r.activities...)
	}

	return buildReport(opts.MonthLabel, opts.Since, opts.Until, repoSummaries, allActivities), nil
}

func collectRepo(ctx context.Context, client *ghClient, repo string, since, until time.Time) (RepoSummary, []SampleActivity, error) {
	commits, err := client.listCommits(ctx, repo, since, until)
	if err != nil {
		return RepoSummary{}, nil, fmt.Errorf("listing commits for %q: %w", repo, err)
	}

	summary := RepoSummary{
		Repo:         repo,
		TotalCommits: len(commits),
	}

	// Process commits in chronological order (oldest first) so a sample file
	// created earlier in the month and edited later in the same month remains
	// classified as "created".
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Commit.Author.Date.Before(commits[j].Commit.Author.Date)
	})

	var candidateSHAs []string
	for _, c := range commits {
		authorLogin := c.Author.Login
		if authorLogin == "" {
			authorLogin = c.Commit.Author.Name
		}
		subject := firstLine(c.Commit.Message)
		if IsDependencyOrBotCommit(authorLogin, subject) {
			summary.SkippedBotOrDepCommits++
			continue
		}
		candidateSHAs = append(candidateSHAs, c.SHA)
	}

	// Fetch commit details with bounded concurrency.
	details := make([]*ghCommitDetail, len(candidateSHAs))
	errs := make([]error, len(candidateSHAs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 6)
	for i, sha := range candidateSHAs {
		wg.Add(1)
		go func(idx int, commitSHA string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			detail, fetchErr := client.getCommit(ctx, repo, commitSHA)
			details[idx] = detail
			errs[idx] = fetchErr
		}(i, sha)
	}
	wg.Wait()

	for i, fetchErr := range errs {
		if fetchErr != nil {
			return RepoSummary{}, nil, fmt.Errorf("fetching commit %s in %s: %w", candidateSHAs[i], repo, fetchErr)
		}
	}

	byFile := make(map[string]*SampleActivity)
	var fileOrder []string

	for _, d := range details {
		author := d.Author.Login
		if author == "" {
			author = d.Commit.Author.Name
		}
		ref := CommitRef{
			SHA:     shortSHA(d.SHA),
			Subject: firstLine(d.Commit.Message),
			Author:  author,
			Date:    d.Commit.Author.Date,
			HTMLURL: d.HTMLURL,
		}

		for _, f := range d.Files {
			if f.Status == "renamed" && f.PreviousFilename != "" {
				if prev, hadPrev := byFile[f.PreviousFilename]; hadPrev {
					delete(byFile, f.PreviousFilename)
					prev.FilePath = f.Filename
					prev.HTMLURL = fmt.Sprintf("https://github.com/%s/blob/main/%s", repo, f.Filename)
					byFile[f.Filename] = prev
					for idx, p := range fileOrder {
						if p == f.PreviousFilename {
							fileOrder[idx] = f.Filename
							break
						}
					}
				}
			}

			change, tags, ok := ClassifyFileChange(repo, f.Filename, f.Status, f.Patch)
			if !ok {
				continue
			}

			existing, seen := byFile[f.Filename]
			if !seen {
				lang := InferLanguage(repo, f.Filename)
				prod := InferProduct(repo, f.Filename)
				existing = &SampleActivity{
					Repo:       repo,
					FilePath:   f.Filename,
					Product:    prod,
					Language:   lang,
					Change:     change,
					RegionTags: tags,
					HTMLURL:    fmt.Sprintf("https://github.com/%s/blob/main/%s", repo, f.Filename),
				}
				byFile[f.Filename] = existing
				fileOrder = append(fileOrder, f.Filename)
			} else {
				// Preserve "created" if the file was created earlier in the window.
				if existing.Change != ChangeCreated && change == ChangeCreated {
					existing.Change = ChangeCreated
				}
				existing.RegionTags = mergeUniqueStrings(existing.RegionTags, tags)
			}
			existing.Commits = append(existing.Commits, ref)
		}
	}

	var activities []SampleActivity
	for _, path := range fileOrder {
		act := *byFile[path]
		if act.Change == ChangeCreated {
			summary.CreatedSamples++
		} else {
			summary.UpdatedSamples++
		}
		activities = append(activities, act)
	}

	return summary, activities, nil
}

// IsDependencyOrBotCommit reports whether a commit was authored by an automated
// dependency bot (such as Dependabot or Renovate) or has a commit message
// indicating a pure dependency version bump.
func IsDependencyOrBotCommit(author, subject string) bool {
	lowerAuthor := strings.ToLower(strings.TrimSpace(author))
	for _, bot := range []string{
		"dependabot",
		"renovate",
		"snyk-bot",
		"google-oss-bot",
		"gcf-owl-bot",
		"owl-bot",
	} {
		if strings.Contains(lowerAuthor, bot) {
			return true
		}
	}

	lowerSubject := strings.ToLower(strings.TrimSpace(subject))
	cleanedSubject := strings.ToLower(cleanCommitSubject(subject))
	for _, candidate := range []string{lowerSubject, cleanedSubject} {
		for _, prefix := range []string{
			"chore(deps",
			"build(deps",
			"ci(deps",
			"fix(deps",
			"deps(",
			"deps:",
			"update dependency ",
			"update dependencies",
			"update package dependencies",
			"update deps",
			"update docker.io/",
			"update mirror.gcr.io/",
			"update eclipse-temurin",
			"update actions/",
		} {
			if strings.HasPrefix(candidate, prefix) {
				return true
			}
		}
	}

	if strings.Contains(lowerSubject, "bump ") &&
		(strings.Contains(lowerSubject, " in /") ||
			strings.Contains(lowerSubject, " in requirements") ||
			strings.Contains(lowerSubject, " from ")) {
		return true
	}

	return false
}

// IsSampleSourceFile reports whether filename is a candidate sample code file
// (excluding build manifests, CI configs, fixtures, and unit/integration test files).
func IsSampleSourceFile(filename string) bool {
	clean := filepath.ToSlash(strings.TrimSpace(filename))
	if clean == "" {
		return false
	}

	for _, ignoredPrefix := range []string{
		".github/",
		".kokoro/",
		".maintenance/",
		"sampletools/",
		"scripts/",
		"testing/",
		"test/",
		"tests/",
	} {
		if strings.HasPrefix(clean, ignoredPrefix) {
			return false
		}
	}

	lowerClean := strings.ToLower(clean)
	for _, ignoredSegment := range []string{
		"/.samples/",
		"/.github/",
		"/.kokoro/",
		"/node_modules/",
		"/vendor/",
		"/test/",
		"/tests/",
		"/__tests__/",
		".tests/",
	} {
		if strings.Contains(lowerClean, ignoredSegment) {
			return false
		}
	}

	base := filepath.Base(clean)
	lowerBase := strings.ToLower(base)

	// Exclude common testrunner/config files and unit/integration test files.
	if lowerBase == "noxfile.py" || lowerBase == "noxfile_config.py" || lowerBase == "conftest.py" || lowerBase == "setup.py" {
		return false
	}
	for _, testSuffix := range []string{
		"_test.go",
		"_test.py",
		"test.java",
		"it.java",
		"tests.cs",
		"test.cs",
		"fixture.cs",
		".test.js",
		".spec.js",
		".test.ts",
		".spec.ts",
		"test.php",
		"_test.rb",
		"_spec.rb",
	} {
		if strings.HasSuffix(lowerBase, testSuffix) {
			return false
		}
	}
	if strings.HasPrefix(lowerBase, "test_") && strings.HasSuffix(lowerBase, ".py") {
		return false
	}

	ext := strings.ToLower(filepath.Ext(base))
	switch ext {
	case ".py", ".js", ".mjs", ".cjs", ".ts", ".java", ".kt", ".go", ".cs", ".php", ".rb", ".tf":
		return true
	default:
		return false
	}
}

// ClassifyFileChange inspects a file status and unified diff patch to determine
// whether it represents a newly created sample, an updated sample, or a
// non-sample change.
func ClassifyFileChange(repo, filename, status, patch string) (ChangeType, []string, bool) {
	if !IsSampleSourceFile(filename) {
		return "", nil, false
	}

	switch status {
	case "removed":
		return "", nil, false
	case "added":
		addedTags, _, _ := extractPatchRegionTags(patch)
		// In client-library-samples, every sample file under <product>/<version>/<lang>/
		// is a dedicated sample entrypoint even if region tags are in a separate commit.
		// In language monorepos, require either at least one region tag or a non-empty patch.
		if len(addedTags) == 0 && !strings.HasSuffix(repo, "/client-library-samples") {
			if !strings.Contains(patch, "[START ") {
				return "", nil, false
			}
		}
		return ChangeCreated, addedTags, true

	case "modified", "renamed":
		if strings.TrimSpace(patch) == "" {
			return "", nil, false
		}
		addedTags, removedTags, allTags := extractPatchRegionTags(patch)
		hasCodeChange := patchHasExecutableChanges(patch)

		// Check if a genuinely new region tag was introduced alongside executable code.
		removedSet := make(map[string]bool, len(removedTags))
		for _, t := range removedTags {
			removedSet[t] = true
		}
		var netNewTags []string
		for _, t := range addedTags {
			if !removedSet[t] {
				netNewTags = append(netNewTags, t)
			}
		}

		if len(netNewTags) > 0 && hasCodeChange && !strings.HasSuffix(repo, "/client-library-samples") {
			return ChangeCreated, netNewTags, true
		}
		if !hasCodeChange {
			return "", nil, false
		}
		return ChangeUpdated, allTags, true

	default:
		return "", nil, false
	}
}

func extractPatchRegionTags(patch string) (added, removed, all []string) {
	lines := strings.Split(patch, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "@@") {
			continue
		}
		matches := regionTagStartRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			tag := m[1]
			if ignoredWrapperTags[strings.ToLower(tag)] {
				continue
			}
			all = appendUnique(all, tag)
			if strings.HasPrefix(line, "+") {
				added = appendUnique(added, tag)
			} else if strings.HasPrefix(line, "-") {
				removed = appendUnique(removed, tag)
			}
		}
	}
	return added, removed, all
}

// patchHasExecutableChanges reports whether the unified diff contains added or
// removed lines other than pure region-tag comments or blank lines.
func patchHasExecutableChanges(patch string) bool {
	for _, line := range strings.Split(patch, "\n") {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "@@") {
			continue
		}
		if !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") {
			continue
		}
		trimmed := strings.TrimSpace(line[1:])
		if trimmed == "" {
			continue
		}
		if regionTagStartRe.MatchString(trimmed) || regionTagEndRe.MatchString(trimmed) {
			continue
		}
		return true
	}
	return false
}

// InferLanguage determines the programming language of a sample from its
// repository and file path.
func InferLanguage(repo, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".py":
		return "Python"
	case ".js", ".mjs", ".cjs", ".ts":
		return "Node.js"
	case ".java":
		return "Java"
	case ".kt":
		return "Kotlin"
	case ".go":
		return "Go"
	case ".cs":
		return "C#"
	case ".php":
		return "PHP"
	case ".rb":
		return "Ruby"
	case ".tf":
		return "Terraform"
	}

	switch {
	case strings.Contains(repo, "python"):
		return "Python"
	case strings.Contains(repo, "nodejs"):
		return "Node.js"
	case strings.Contains(repo, "java"):
		return "Java"
	case strings.Contains(repo, "golang"):
		return "Go"
	case strings.Contains(repo, "dotnet"):
		return "C#"
	case strings.Contains(repo, "php"):
		return "PHP"
	case strings.Contains(repo, "ruby"):
		return "Ruby"
	default:
		return "Other"
	}
}

// InferProduct maps a repository and sample file path to a human-readable
// Google Cloud product or service name.
func InferProduct(repo, filename string) string {
	clean := filepath.ToSlash(strings.TrimPrefix(filename, "/"))
	parts := strings.Split(clean, "/")
	if len(parts) == 0 {
		return "General"
	}
	slug := parts[0]
	if name, ok := productDisplayNames[slug]; ok {
		return name
	}
	if name, ok := productDisplayNames[strings.ToLower(slug)]; ok {
		return name
	}
	// Format unknown slug into title case.
	words := strings.FieldsFunc(slug, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	if len(words) == 0 {
		return slug
	}
	return strings.Join(words, " ")
}

func buildReport(monthLabel string, since, until time.Time, repos []RepoSummary, activities []SampleActivity) *Report {
	prodMap := make(map[string]*ProductSummary)
	langSet := make(map[string]bool)
	totalCreated := 0
	totalUpdated := 0

	for _, act := range activities {
		langSet[act.Language] = true
		ps, ok := prodMap[act.Product]
		if !ok {
			ps = &ProductSummary{Product: act.Product}
			prodMap[act.Product] = ps
		}
		ps.Languages = appendUnique(ps.Languages, act.Language)
		if act.Change == ChangeCreated {
			ps.Created = append(ps.Created, act)
			totalCreated++
		} else {
			ps.Updated = append(ps.Updated, act)
			totalUpdated++
		}
	}

	var products []ProductSummary
	for _, ps := range prodMap {
		sort.Strings(ps.Languages)
		products = append(products, *ps)
	}
	sort.Slice(products, func(i, j int) bool {
		// Sort products with the most created samples first, then by name.
		if len(products[i].Created) != len(products[j].Created) {
			return len(products[i].Created) > len(products[j].Created)
		}
		if len(products[i].Updated) != len(products[j].Updated) {
			return len(products[i].Updated) > len(products[j].Updated)
		}
		return products[i].Product < products[j].Product
	})

	var langs []string
	for l := range langSet {
		langs = append(langs, l)
	}
	sort.Strings(langs)

	return &Report{
		MonthLabel:   monthLabel,
		Since:        since,
		Until:        until,
		TotalCreated: totalCreated,
		TotalUpdated: totalUpdated,
		Languages:    langs,
		Repos:        repos,
		Products:     products,
	}
}

type ghClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type ghCommitSummary struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string    `json:"name"`
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
}

type ghCommitDetail struct {
	SHA     string `json:"sha"`
	HTMLURL string `json:"html_url"`
	Commit  struct {
		Message string `json:"message"`
		Author  struct {
			Name string    `json:"name"`
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	Files []struct {
		Filename         string `json:"filename"`
		PreviousFilename string `json:"previous_filename"`
		Status           string `json:"status"`
		Patch            string `json:"patch"`
	} `json:"files"`
}

func (c *ghClient) listCommits(ctx context.Context, repo string, since, until time.Time) ([]ghCommitSummary, error) {
	var all []ghCommitSummary
	for page := 1; page <= 10; page++ {
		u := fmt.Sprintf(
			"%s/repos/%s/commits?since=%s&until=%s&per_page=100&page=%d",
			c.baseURL,
			repo,
			url.QueryEscape(since.UTC().Format(time.RFC3339)),
			url.QueryEscape(until.UTC().Format(time.RFC3339)),
			page,
		)
		var batch []ghCommitSummary
		if err := c.getJSON(ctx, u, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

func (c *ghClient) getCommit(ctx context.Context, repo, sha string) (*ghCommitDetail, error) {
	u := fmt.Sprintf("%s/repos/%s/commits/%s", c.baseURL, repo, url.PathEscape(sha))
	var detail ghCommitDetail
	if err := c.getJSON(ctx, u, &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func (c *ghClient) getJSON(ctx context.Context, rawURL string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, rawURL, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decoding response from %s: %w", rawURL, err)
	}
	return nil
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return strings.TrimSpace(s)
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func appendUnique(slice []string, val string) []string {
	for _, existing := range slice {
		if existing == val {
			return slice
		}
	}
	return append(slice, val)
}

func mergeUniqueStrings(a, b []string) []string {
	out := append([]string(nil), a...)
	for _, item := range b {
		out = appendUnique(out, item)
	}
	return out
}
