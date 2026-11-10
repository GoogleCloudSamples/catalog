# Use case 1: The API is enabled when its operation completes, but needs time
# to propagate before other resources can use it.
resource "google_project_service" "workflows" {
  project            = var.project_id
  service            = "workflows.googleapis.com"
  disable_on_destroy = false
}

# Wait a fixed duration for the newly enabled API to propagate.
resource "time_sleep" "wait_for_api" {
  create_duration = "30s"

  depends_on = [google_project_service.workflows]
}

resource "google_workflows_workflow" "default" {
  project         = var.project_id
  name            = "example-workflow"
  region          = var.location_id
  source_contents = "- done:\n    return: ok\n"

  depends_on = [time_sleep.wait_for_api]
}

# Use case 2: The instance exists when its operation completes, but its web
# server isn't listening until the startup script finishes.
resource "google_compute_instance" "web" {
  project      = var.project_id
  name         = "example-web-server"
  machine_type = "e2-micro"
  zone         = "${var.location_id}-a"
  tags         = ["example-web-server"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    network = "default"
    access_config {}
  }

  metadata_startup_script = "apt-get update && apt-get install -y nginx"
}

# Allow HTTP only from the address that runs Terraform.
resource "google_compute_firewall" "allow_http" {
  project       = var.project_id
  name          = "example-allow-http"
  network       = "default"
  source_ranges = [var.client_cidr]
  target_tags   = ["example-web-server"]

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
}

# Probe the server until it responds. The probe retries connection errors and
# 5xx responses for up to 60 seconds (12 attempts, 5 seconds apart).
data "http" "wait_for_server" {
  url = "http://${google_compute_instance.web.network_interface[0].access_config[0].nat_ip}/"

  retry {
    attempts     = 12
    min_delay_ms = 5000
  }

  depends_on = [google_compute_firewall.allow_http]
}
