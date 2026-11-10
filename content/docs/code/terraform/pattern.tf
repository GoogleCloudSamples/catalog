resource "google_project_service" "default" {
  service            = "cloudkms.googleapis.com"
  disable_on_destroy = false
}

resource "google_kms_key_ring" "default" {
  name     = "example-key-ring"
  location = "us-central1"

  depends_on = [google_project_service.default]
}

resource "google_kms_crypto_key" "default" {
  name     = "example-crypto-key"
  key_ring = google_kms_key_ring.default.id
  purpose  = "ENCRYPT_DECRYPT"

  # Rotate keys every 90 days (7,776,000 seconds) to limit exposure.
  rotation_period = "7776000s"
}
