package metrics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIsDependencyOrBotCommit(t *testing.T) {
	tests := []struct {
		name    string
		author  string
		subject string
		want    bool
	}{
		{
			name:    "dependabot_author",
			author:  "dependabot[bot]",
			subject: "Bump urllib3 from 2.0.0 to 2.1.0",
			want:    true,
		},
		{
			name:    "renovate_author",
			author:  "renovate-bot",
			subject: "Update dependency google-cloud-asset to v4 (#14590)",
			want:    true,
		},
		{
			name:    "human chore deps commit",
			author:  "XrossFox",
			subject: "chore(deps): update transformers and dependencies in dataflow/run-inference (#14608)",
			want:    true,
		},
		{
			name:    "human build deps bump",
			author:  "XrossFox",
			subject: "build(multiple samples): bump Django to 6.1.1 in requirements (#14575)",
			want:    true,
		},
		{
			name:    "feature commit adding samples",
			author:  "iennae",
			subject: "feat(bigqueryreservation): add Java v1 samples and fixtures (#138)",
			want:    false,
		},
		{
			name:    "fix commit updating sample code",
			author:  "d831f8b-author",
			subject: "fix(genai): update 5 samples to gemini-3.8-flash (#14676)",
			want:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsDependencyOrBotCommit(tc.author, tc.subject); got != tc.want {
				t.Errorf("IsDependencyOrBotCommit(%q, %q) = %v, want %v", tc.author, tc.subject, got, tc.want)
			}
		})
	}
}

func TestIsSampleSourceFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"bigquerydatatransfer/v1/java/DataTransferConfigCreate.java", true},
		{"storage/snippets/upload_file.py", true},
		{"storage/snippets/upload_file_test.py", false},
		{"storage/snippets/test_upload_file.py", false},
		{"storage/src/test/java/com/example/StorageIT.java", false},
		{"bigquerydatatransfer/.samples/spec.json", false},
		{".maintenance/fixtures/main.tf", false},
		{"sampletools/validators/audit.py", false},
		{"noxfile_config.py", false},
		{"package.json", false},
	}

	for _, tc := range tests {
		if got := IsSampleSourceFile(tc.filename); got != tc.want {
			t.Errorf("IsSampleSourceFile(%q) = %v, want %v", tc.filename, got, tc.want)
		}
	}
}

func TestClassifyFileChange(t *testing.T) {
	tests := []struct {
		name       string
		repo       string
		filename   string
		status     string
		patch      string
		wantChange ChangeType
		wantTags   []string
		wantOK     bool
	}{
		{
			name:     "added file with region tag",
			repo:     "GoogleCloudPlatform/python-docs-samples",
			filename: "storagecontrol/delete_folder_recursive.py",
			status:   "added",
			patch: `@@ -0,0 +1,10 @@
+# [START storage_control_delete_folder_recursive]
+def delete_folder_recursive(bucket_name: str) -> None:
+    print(bucket_name)
+# [END storage_control_delete_folder_recursive]`,
			wantChange: ChangeCreated,
			wantTags:   []string{"storage_control_delete_folder_recursive"},
			wantOK:     true,
		},
		{
			name:     "modified file with only region tag comment added",
			repo:     "GoogleCloudSamples/client-library-samples",
			filename: "bigquerystorage/v1/python/read_rows.py",
			status:   "modified",
			patch: `@@ -1,3 +1,5 @@
 # [START bigquerystorage_v1_read_rows]
+# [START bigquerystorage_read_rows]
 def read_rows():
     pass`,
			wantChange: "",
			wantTags:   nil,
			wantOK:     false,
		},
		{
			name:     "modified file with executable code updated",
			repo:     "GoogleCloudPlatform/python-docs-samples",
			filename: "genai/text_generation/text_gen_with_txt.py",
			status:   "modified",
			patch: `@@ -5,4 +5,4 @@
 # [START googlegenaisdk_textgen_with_txt]
-model = "gemini-2.5-flash"
+model = "gemini-3.8-flash"
 # [END googlegenaisdk_textgen_with_txt]`,
			wantChange: ChangeUpdated,
			wantTags:   []string{"googlegenaisdk_textgen_with_txt"},
			wantOK:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotChange, gotTags, gotOK := ClassifyFileChange(tc.repo, tc.filename, tc.status, tc.patch)
			if gotOK != tc.wantOK || gotChange != tc.wantChange {
				t.Fatalf("ClassifyFileChange() = (%q, %v, %v), want (%q, %v, %v)", gotChange, gotTags, gotOK, tc.wantChange, tc.wantTags, tc.wantOK)
			}
			if strings.Join(gotTags, ",") != strings.Join(tc.wantTags, ",") {
				t.Errorf("ClassifyFileChange() tags = %v, want %v", gotTags, tc.wantTags)
			}
		})
	}
}

func TestCollectAndWriteMarkdown(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/GoogleCloudSamples/client-library-samples/commits", func(w http.ResponseWriter, r *http.Request) {
		commits := []ghCommitSummary{
			{
				SHA: "333333333333",
				Commit: struct {
					Message string `json:"message"`
					Author  struct {
						Name string    `json:"name"`
						Date time.Time `json:"date"`
					} `json:"author"`
				}{
					Message: "fix(bigqueryreservation): refine Java reservation sample (#139)",
				},
			},
			{
				SHA: "222222222222",
				Commit: struct {
					Message string `json:"message"`
					Author  struct {
						Name string    `json:"name"`
						Date time.Time `json:"date"`
					} `json:"author"`
				}{
					Message: "chore(deps): bump js-yaml (#124)",
				},
			},
			{
				SHA: "111111111111",
				Commit: struct {
					Message string `json:"message"`
					Author  struct {
						Name string    `json:"name"`
						Date time.Time `json:"date"`
					} `json:"author"`
				}{
					Message: "feat(bigqueryreservation): add Java v1 samples (#138)",
				},
			},
		}
		commits[0].Commit.Author.Date = time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
		commits[0].Author.Login = "iennae"
		commits[1].Commit.Author.Date = time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
		commits[1].Author.Login = "dependabot[bot]"
		commits[2].Commit.Author.Date = time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
		commits[2].Author.Login = "iennae"
		_ = json.NewEncoder(w).Encode(commits)
	})

	mux.HandleFunc("/repos/GoogleCloudSamples/client-library-samples/commits/111111111111", func(w http.ResponseWriter, r *http.Request) {
		detail := ghCommitDetail{
			SHA:     "111111111111",
			HTMLURL: "https://github.com/GoogleCloudSamples/client-library-samples/commit/111111111111",
		}
		detail.Author.Login = "iennae"
		detail.Commit.Message = "feat(bigqueryreservation): add Java v1 samples (#138)"
		detail.Files = []struct {
			Filename         string `json:"filename"`
			PreviousFilename string `json:"previous_filename"`
			Status           string `json:"status"`
			Patch            string `json:"patch"`
		}{
			{
				Filename: "bigqueryreservation/v1/java/ReservationCreate.java",
				Status:   "added",
				Patch:    "+// [START bigqueryreservation_v1_reservation_create]\n+class ReservationCreate {}\n+// [END bigqueryreservation_v1_reservation_create]",
			},
		}
		_ = json.NewEncoder(w).Encode(detail)
	})

	mux.HandleFunc("/repos/GoogleCloudSamples/client-library-samples/commits/333333333333", func(w http.ResponseWriter, r *http.Request) {
		detail := ghCommitDetail{
			SHA:     "333333333333",
			HTMLURL: "https://github.com/GoogleCloudSamples/client-library-samples/commit/333333333333",
		}
		detail.Author.Login = "iennae"
		detail.Commit.Message = "fix(bigqueryreservation): refine Java reservation sample (#139)"
		detail.Files = []struct {
			Filename         string `json:"filename"`
			PreviousFilename string `json:"previous_filename"`
			Status           string `json:"status"`
			Patch            string `json:"patch"`
		}{
			{
				Filename:         "bigqueryreservation/v1/java/CreateReservation.java",
				PreviousFilename: "bigqueryreservation/v1/java/ReservationCreate.java",
				Status:           "renamed",
				Patch:            "-class ReservationCreate {}\n+public class CreateReservation {}",
			},
		}
		_ = json.NewEncoder(w).Encode(detail)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	since, until, err := ParseMonthWindow("2026-10")
	if err != nil {
		t.Fatalf("ParseMonthWindow() unexpected error: %v", err)
	}

	rep, err := Collect(t.Context(), Options{
		Since:      since,
		Until:      until,
		MonthLabel: "2026-10",
		Repos:      []string{"GoogleCloudSamples/client-library-samples"},
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	if rep.TotalCreated != 1 || rep.TotalUpdated != 0 {
		t.Fatalf("Collect() totals = (created=%d, updated=%d), want (1, 0)", rep.TotalCreated, rep.TotalUpdated)
	}
	if rep.Repos[0].SkippedBotOrDepCommits != 1 {
		t.Errorf("SkippedBotOrDepCommits = %d, want 1", rep.Repos[0].SkippedBotOrDepCommits)
	}

	var md bytes.Buffer
	if err := WriteMarkdown(&md, rep); err != nil {
		t.Fatalf("WriteMarkdown() unexpected error: %v", err)
	}
	out := md.String()
	if !strings.Contains(out, "**BigQuery Reservation**") || !strings.Contains(out, "1 new sample") {
		t.Errorf("WriteMarkdown() output missing expected product summary:\n%s", out)
	}
}
