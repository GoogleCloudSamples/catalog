resource "google_storage_bucket" "default" {
  name                        = "example-archive-bucket"
  location                    = "US"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  force_destroy               = true

  # Retain soft-deleted objects for 7 days (604,800 seconds) to protect against accidental deletion.
  soft_delete_policy {
    retention_duration_seconds = 604800
  }
}
