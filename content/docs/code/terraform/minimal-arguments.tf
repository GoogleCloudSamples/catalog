variable "project_id" {
  type        = string
  description = "The Google Cloud project ID, for example \"example-project-id\""
}

variable "location_id" {
  type        = string
  description = "The location ID, for example \"us-central1\""
}

variable "resource_id" {
  type        = string
  description = "The ID of an existing resource, for example \"example-resource-42\""
}

variable "local_file_path" {
  type        = string
  description = "The path to a local input file, for example \"/tmp/example-input.txt\""
}
