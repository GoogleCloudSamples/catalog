# [START sandwichmanager_make_sandwich]
resource "google_project_iam_member" "default" {
  project = data.google_project.default.id
  role    = "roles/viewer"
  member  = "user:jane@example.com"
}
# [END sandwichmanager_make_sandwich]
