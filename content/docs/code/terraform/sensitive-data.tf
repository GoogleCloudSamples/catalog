resource "google_kms_secret_ciphertext" "encrypted" {
  crypto_key = var.crypto_key_id
  plaintext  = "Hello, world!"
}

# Use sensitive data without printing it.
output "ciphertext" {
  value     = google_kms_secret_ciphertext.encrypted.ciphertext
  sensitive = true
}
