# Manage a single binding so existing project IAM bindings are preserved.
resource "google_project_iam_member" "viewer" {
  project = var.project_id
  role    = "roles/viewer"
  member  = "user:jane@example.com"
}
